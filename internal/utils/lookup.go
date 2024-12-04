package utils

import (
	"github.com/sem-hub/ygg-peers-select/internal/parse"
)

func FqdnLookup(list *[]parse.Peer, key string) []string {

	for _, peer := range *list {
		for _, ip := range peer.IpList {
			if ip == key {
				return peer.Uris
			}
		}
	}
	return []string{}
}
