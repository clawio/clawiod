package cmd

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login <username> <password>",
	Short: "Login into ClawIO",
	Run:   login,
}

func login(cmd *cobra.Command, args []string) {

	if len(args) != 2 {
		cmd.Help()
		os.Exit(1)
	}

	logger := getLogger()
	client := http.DefaultClient

	tokenReq := &tokenReq{Username: args[0], Password: args[1]}
	jsonBody, err := json.Marshal(tokenReq)
	if err != nil {
		logger.Error().Log("error", err, "msg", "error encoding token request")
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", os.Getenv("CLAWIOBENCH_AUTH_ADDR")+"/token", bytes.NewReader(jsonBody))
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	if res.StatusCode == http.StatusCreated {
		jsonRes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			logger.Error().Log("error", err)
			os.Exit(1)
		}

		tokenRes := &tokenRes{}
		err = json.Unmarshal(jsonRes, tokenRes)
		if err != nil {
			logger.Error().Log("error", err)
			os.Exit(1)
		}

		// Save token into $HOME/.clawiobench/credentials
		u, err := user.Current()
		if err != nil {
			logger.Error().Log("error", err)
			os.Exit(1)
		}

		err = os.MkdirAll(path.Join(u.HomeDir, ".clawiobench"), 0755)
		if err != nil {
			logger.Error().Log("error", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile(path.Join(u.HomeDir, ".clawiobench", "credentials"), []byte(tokenRes.AccessToken), 0644)
		if err != nil {
			logger.Error().Log("error", err)
			os.Exit(1)
		}
		os.Exit(0)

	}

	if res.StatusCode == http.StatusUnauthorized {
		logger.Error().Log("error", "unauthorized")
		os.Exit(1)
	}
	if res.StatusCode == http.StatusBadRequest {
		logger.Error().Log("error", "unauthorized")
		os.Exit(1)
	}

	logger.Error().Log("error", "unauthorized")
	os.Exit(1)
}

func init() {
	RootCmd.AddCommand(loginCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

type tokenReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenRes struct {
	AccessToken string `json:"access_token"`
}
*/
