package main

import (
	"net/http"

	"github.com/cowry-network/go/clients/horizon"
	"github.com/cowry-network/go/services/friendbot/internal"
	"github.com/cowry-network/go/strkey"
)

func initFriendbot(
	friendbotSecret string,
	networkPassphrase string,
	horizonURL string,
	startingBalance string,
) *internal.Bot {

	if friendbotSecret == "" || networkPassphrase == "" || horizonURL == "" || startingBalance == "" {
		return nil
	}

	// ensure its a seed if its not blank
	strkey.MustDecode(strkey.VersionByteSeed, friendbotSecret)

	return &internal.Bot{
		Secret: friendbotSecret,
		Horizon: &horizon.Client{
			URL:  horizonURL,
			HTTP: http.DefaultClient,
		},
		Network:           networkPassphrase,
		StartingBalance:   startingBalance,
		SubmitTransaction: internal.AsyncSubmitTransaction,
	}
}
