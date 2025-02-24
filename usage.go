package clingy

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

func (env *Environment) printUsage(ctx context.Context, st *runState, desc cmdDesc) {
	tw := tabwriter.NewWriter(env.Stdout, 4, 4, 4, ' ', 0)
	defer tw.Flush()

	printErrors(ctx, tw, st.errors)
	printUsagePrefix(ctx, tw, st, desc)
	printSubcommands(ctx, tw, st, desc.subcmds)
	printArguments(ctx, tw, st.pos)
	printFlags(ctx, tw, st)
	printGlobalFlags(ctx, tw, st)
	printUsageSuffix(ctx, tw, st, len(desc.subcmds) > 0)
}

func printErrors(ctx context.Context, w io.Writer, errs []error) {
	if len(errs) == 0 {
		return
	}

	fmt.Fprintln(w, "Errors:")
	for _, err := range errs {
		fmt.Fprintf(w, "\t%s\n", err)
	}
	fmt.Fprintln(w)
}

func printUsagePrefix(ctx context.Context, w io.Writer, st *runState, desc cmdDesc) {
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "\t%s", st.name())

	if st.advanced {
		st.flags.params(func(p *param) {
			if p == nil {
				return
			} else if p.rep {
				fmt.Fprintf(w, " [--%s %s ...]", p.name, p.flagType())
			} else if typ := p.flagType(); typ != "" {
				fmt.Fprintf(w, " [--%s %s]", p.name, p.flagType())
			} else {
				fmt.Fprintf(w, " [--%s]", p.name)
			}
		})
	} else if st.flags.getCount() > 0 {
		fmt.Fprintf(w, " [flags]")
	}

	optionals := 0
	st.pos.params(func(p *param) {
		switch {
		case p.rep:
			fmt.Fprintf(w, " [%s ...]", p.name)
		case p.opt:
			fmt.Fprintf(w, " [%s", p.name)
			optionals++
		default:
			fmt.Fprintf(w, " <%s>", p.name)
		}
	})
	for i := 0; i < optionals; i++ {
		fmt.Fprint(w, "]")
	}

	if len(desc.subcmds) > 0 {
		fmt.Fprint(w, " [command]")
	}
	fmt.Fprintln(w)

	if len(desc.short) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "\t"+desc.short)
	}
	if len(desc.long) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "\t"+strings.Join(strings.Split(desc.long, "\n"), "\n\t"))
	}
}

func printSubcommands(ctx context.Context, w io.Writer, st *runState, descs []cmdDesc) {
	hp := newHeaderPrinter(w, "Available commands:")
	for _, desc := range descs {
		fmt.Fprintf(hp, "\t%s\t%s\n", desc.name, desc.short)
	}
}

func printArguments(ctx context.Context, w io.Writer, pos *paramsPos) {
	hp := newHeaderPrinter(w, "Arguments:")
	pos.params(func(p *param) {
		fmt.Fprintf(hp, "\t%s\t%s\n", p.name, p.desc)
	})
}

func printFlags(ctx context.Context, w io.Writer, st *runState) {
	hp := newHeaderPrinter(w, "Flags:")
	st.flags.params(func(p *param) {
		if st.advanced || p == nil || !p.adv {
			printFlag(ctx, hp, p)
		}
	})
}

func printGlobalFlags(ctx context.Context, w io.Writer, st *runState) {
	hp := newHeaderPrinter(w, "Global flags:")
	st.gflags.params(func(p *param) {
		if st.advanced || p == nil || !p.adv {
			printFlag(ctx, hp, p)
		}
	})
}

func printFlag(ctx context.Context, w io.Writer, p *param) {
	if p == nil {
		fmt.Fprintln(w)
		return
	}
	fmt.Fprint(w, "\t")
	if p.short != 0 {
		fmt.Fprintf(w, "-%c, ", p.short)
	} else {
		fmt.Fprint(w, "    ")
	}
	fmt.Fprintf(w, "--%s %s\t%s", p.name, p.flagType(), p.desc)
	if p.rep {
		fmt.Fprintf(w, " (repeated)")
	}
	if !isZero(p.def) {
		fmt.Fprintf(w, " (default %#v)", p.def)
	}
	fmt.Fprintln(w)
}

func printUsageSuffix(ctx context.Context, w io.Writer, st *runState, subcmds bool) {
	if subcmds {
		fmt.Fprintf(w, "\nUse \"%s [command] --help\" for more information about a command.\n", st.name())
	}
}

func isZero(x interface{}) bool {
	rv := reflect.ValueOf(x)
	return !rv.IsValid() || rv.IsZero() || (rv.Kind() == reflect.Slice && rv.Len() == 0)
}
