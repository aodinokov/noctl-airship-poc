// Package main implements pod emulation function to run arbitrary scripts and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/crypter"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

type CryptCmdParams struct {
	Password string
	Value    string
}

func NewCryptCmd(name string) *cobra.Command {
	p := CryptCmdParams{}
	c := &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("%s field with the password provided in the arg", name),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("called %s", cmd.Use)

			key, err := crypter.Key(p.Password)
			if err != nil {
				return fmt.Errorf("wasn't able to get key from password: %v", err)
			}
			//log.Printf("pass: %s key: %v", p.Password, key)

			result := ""

			switch cmd.Use {
			case "encrypt":
				result, err = crypter.Encrypt(p.Value, key)
			case "decrypt":
				result, err = crypter.Decrypt(p.Value, key)
			}

			if err != nil {
				return fmt.Errorf("operation failed: %v", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", result)
			return nil
		},
	}

	c.Flags().StringVar(&p.Password, "password", "", "password used for crypt operation")
	c.Flags().StringVar(&p.Value, "value", "", "value for crypt operation")

	return c
}

func main() {
	log.Print("started")
	defer log.Print("Finished")

	cfg := crypter.FunctionConfig{}
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}

	cmd := framework.Command(resourceList, func() error {
		fn, err := crypter.NewFunction(&cfg)
		if err != nil {
			log.Printf("function creation failed: %v", err)
			return err
		}

		items, err := fn.Exec(resourceList.Items)
		if err != nil {
			return err
		}
		resourceList.Items = items
		return nil
	})

	// additional features
	en := NewCryptCmd("encrypt")
	de := NewCryptCmd("decrypt")

	cmd.AddCommand(en, de)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
