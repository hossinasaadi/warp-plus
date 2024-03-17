package main

import (
	"context"
	"errors"
	"fmt"

	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hossinasaadi/warp-plus/app"
	"github.com/hossinasaadi/warp-plus/warp"
	"golang.org/x/exp/slog"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

var psiphonCountries = []string{
	"AT",
	"BE",
	"BG",
	"BR",
	"CA",
	"CH",
	"CZ",
	"DE",
	"DK",
	"EE",
	"ES",
	"FI",
	"FR",
	"GB",
	"HU",
	"IE",
	"IN",
	"IT",
	"JP",
	"LV",
	"NL",
	"NO",
	"PL",
	"RO",
	"RS",
	"SE",
	"SG",
	"SK",
	"UA",
	"US",
}

func main() {
	fs := ff.NewFlagSet("warp-plus")
	var (
		verbose  = fs.Bool('v', "verbose", "enable verbose logging")
		bind     = fs.String('b', "bind", "127.0.0.1:8086", "socks bind address")
		endpoint = fs.String('e', "endpoint", "", "warp endpoint")
		key      = fs.String('k', "key", "", "warp key")
		country  = fs.StringEnumLong("country", fmt.Sprintf("psiphon country code (valid values: %s)", psiphonCountries), psiphonCountries...)
		psiphon  = fs.BoolLong("cfon", "enable psiphon mode (must provide country as well)")
		gool     = fs.BoolLong("gool", "enable gool mode (warp in warp)")
		scan     = fs.BoolLong("scan", "enable warp scanning (experimental)")
		rtt      = fs.DurationLong("rtt", 1000*time.Millisecond, "scanner rtt limit")
	)

	// Config file and envvars can be added through ff later
	err := ff.Parse(fs, os.Args[1:])
	switch {
	case errors.Is(err, ff.ErrHelp):
		fmt.Fprintf(os.Stderr, "%s\n", ffhelp.Flags(fs))
		os.Exit(0)
	case err != nil:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if *verbose {
		l = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	if *psiphon && *gool {
		fatal(l, errors.New("can't use cfon and gool at the same time"))
	}

	bindAddrPort, err := netip.ParseAddrPort(*bind)
	if err != nil {
		fatal(l, fmt.Errorf("invalid bind address: %w", err))
	}

	opts := app.WarpOptions{
		Bind:     bindAddrPort,
		Endpoint: *endpoint,
		License:  *key,
		Gool:     *gool,
	}

	if *psiphon {
		l.Info("psiphon mode enabled", "country", *country)
		opts.Psiphon = &app.PsiphonOptions{Country: *country}
	}

	if *scan {
		l.Info("scanner mode enabled", "max-rtt", rtt)
		opts.Scan = &app.ScanOptions{MaxRTT: *rtt}
	}

	// If the endpoint is not set, choose a random warp endpoint
	if opts.Endpoint == "" {
		addrPort, err := warp.RandomWarpEndpoint()
		if err != nil {
			fatal(l, err)
		}
		opts.Endpoint = addrPort.String()
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := app.RunWarp(ctx, l, opts); err != nil {
			fatal(l, err)
		}
	}()

	<-ctx.Done()
}

func fatal(l *slog.Logger, err error) {
	l.Error(err.Error())
	os.Exit(1)
}
