package cli

import (
	"testing"

	flag "github.com/spf13/pflag"
)

func parseArgs(args []string) (block, scale int, mode, palette string, logical, ver bool, positional []string, err error) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	b := fs.Int("block", 8, "")
	s := fs.Int("scale", 0, "")
	m := fs.String("mode", "average", "")
	p := fs.String("palette", "original", "")
	l := fs.Bool("logical", false, "")
	v := fs.Bool("version", false, "")
	err = fs.Parse(normalizeArgs(args))
	if err != nil {
		return
	}
	return *b, *s, *m, *p, *l, *v, fs.Args(), nil
}

func TestFlagsBeforeArgs(t *testing.T) {
	block, scale, _, _, _, _, pos, err := parseArgs([]string{"--block", "16", "--scale", "4", "in.png", "out.png"})
	if err != nil {
		t.Fatal(err)
	}
	if block != 16 || scale != 4 {
		t.Fatalf("got block=%d scale=%d", block, scale)
	}
	if len(pos) != 2 || pos[0] != "in.png" || pos[1] != "out.png" {
		t.Fatalf("got positional=%v", pos)
	}
}

func TestFlagsAfterArgs(t *testing.T) {
	block, scale, _, _, _, _, pos, err := parseArgs([]string{"in.png", "out.png", "--block", "8", "--scale", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if block != 8 || scale != 2 {
		t.Fatalf("got block=%d scale=%d", block, scale)
	}
	if len(pos) != 2 || pos[0] != "in.png" || pos[1] != "out.png" {
		t.Fatalf("got positional=%v", pos)
	}
}

func TestFlagsInterspersed(t *testing.T) {
	_, _, _, _, logical, _, pos, err := parseArgs([]string{"in.png", "--logical", "out.png"})
	if err != nil {
		t.Fatal(err)
	}
	if !logical {
		t.Fatal("expected logical=true")
	}
	if len(pos) != 2 || pos[0] != "in.png" || pos[1] != "out.png" {
		t.Fatalf("got positional=%v", pos)
	}
}

func TestVersionFlag(t *testing.T) {
	_, _, _, _, _, ver, _, err := parseArgs([]string{"--version"})
	if err != nil {
		t.Fatal(err)
	}
	if !ver {
		t.Fatal("expected version=true")
	}
}

func TestDefaults(t *testing.T) {
	block, scale, mode, palette, logical, _, _, err := parseArgs([]string{"in.png", "out.png"})
	if err != nil {
		t.Fatal(err)
	}
	if block != 8 || scale != 0 || mode != "average" || palette != "original" || logical {
		t.Fatalf("unexpected defaults: block=%d scale=%d mode=%s palette=%s logical=%v", block, scale, mode, palette, logical)
	}
}

func TestSingleDashLongFlag(t *testing.T) {
	_, scale, _, _, _, _, pos, err := parseArgs([]string{"in.png", "out.png", "-scale", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if scale != 3 {
		t.Fatalf("got scale=%d, want 3", scale)
	}
	if len(pos) != 2 {
		t.Fatalf("got positional=%v", pos)
	}
}
