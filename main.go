package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
	flag "github.com/spf13/pflag"
)

func main() {
	os.Exit(run())
}

func run() int {
	version := flag.Bool("version", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return 0
	}

	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No file specified")
		return 1
	}

	var decrypt bool

	for i, f := range files {
		_, n := path.Split(f)
		if len(n) == 0 || (len(n) == 4 && n == ".age") {
			fmt.Fprintf(os.Stderr, "Invalid file: %s\n", f)
			return 1
		}

		if i == 0 {
			decrypt = strings.HasSuffix(n, ".age")
		} else if decrypt != strings.HasSuffix(n, ".age") {
			fmt.Fprintln(os.Stderr, "Cannot mix encryption and decryption")
			return 1
		}
	}

	for {
		files = doAll(files, decrypt)

		if len(files) == 0 {
			return 0
		}

		fmt.Fprintf(os.Stderr, "Failed to process some files, retrying\n\n")
	}
}

func doAll(files []string, decrypt bool) []string {
	var failed []string

	pass, err := readPassphrase(decrypt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read passphrase: %v\n", err)
		return files
	}

	var id age.Identity
	var rp age.Recipient

	if decrypt {
		id, err = age.NewScryptIdentity(pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create identity: %v\n", err)
			return files
		}
	} else {
		rp, err = age.NewScryptRecipient(pass)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create recipient: %v\n", err)
			return files
		}
	}

	for _, f := range files {
		i, err := os.Open(f)
		if err != nil {
			fmt.Printf("Failed to open: %v\n", err)
			failed = append(failed, f)
			continue
		}
		defer i.Close()

		s, err := i.Stat()
		if err != nil {
			fmt.Printf("Failed to stat: %v\n", err)
			failed = append(failed, f)
			continue
		}

		var of string
		if decrypt {
			of = strings.TrimSuffix(f, ".age")
		} else {
			of = f + ".age"
		}

		o, err := os.OpenFile(of, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, s.Mode())
		if err != nil {
			fmt.Printf("Failed to open: %v\n", err)
			failed = append(failed, f)
			continue
		}
		defer o.Close()

		if decrypt {
			var in io.Reader

			rr := bufio.NewReader(i)
			if h, _ := rr.Peek(len(armor.Header)); string(h) == armor.Header {
				in = armor.NewReader(rr)
			} else {
				in = rr
			}

			r, err := age.Decrypt(in, id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to decrypt %s: %v\n", f, err)
				failed = append(failed, f)
				continue
			}

			if _, err = io.Copy(o, r); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to decrypt %s: %v\n", f, err)
				failed = append(failed, f)
				continue
			}

			fmt.Fprintf(os.Stderr, "Decrypted %s\n", f)
		} else {
			w, err := age.Encrypt(o, rp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encrypt %s: %v\n", f, err)
				failed = append(failed, f)
				continue
			}

			if _, err = io.Copy(w, i); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encrypt %s: %v\n", f, err)
				failed = append(failed, f)
				continue
			}

			if err = w.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to encrypt %s: %v\n", f, err)
				failed = append(failed, f)
				continue
			}

			fmt.Fprintf(os.Stderr, "Encrypted %s\n", f)
		}

		if err := os.Remove(f); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete %s: %v\n", f, err)
		}
	}

	return failed
}

var Version = "dev"
