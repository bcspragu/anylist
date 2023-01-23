package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bcspragu/anylist/anylist"
	"github.com/bcspragu/anylist/pb"
	"github.com/namsral/flag"
	"github.com/rs/cors"
	"go.mozilla.org/sops/v3/decrypt"
)

type SecretConfig struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	RefreshToken string `json:"refresh_token"`
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("args cannot be empty")
	}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		sopsConfigPath  = fs.String("sops_encrypted_config", "secrets.enc.json", "A JSON-formatted configuration file for our main server, parseable by the SOPS tool (https://github.com/mozilla/sops).")
		port            = fs.Int("port", 8080, "The port to serve the  HTTP API service on.")
		groceryListName = fs.String("grocery_list_name", "Grokeries 2.0", "The name of the AnyList list to target.")
	)
	// Allows for passing in configuration via a -config path/to/env-file.conf
	// flag, see https://pkg.go.dev/github.com/namsral/flag#readme-usage
	fs.String(flag.DefaultConfigFlagname, "", "path to config file")
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	secCfg, err := decryptConfig(*sopsConfigPath)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := anylist.New(ctx, secCfg.Email, secCfg.Password)
	// c, err := anylist.FromRefreshToken(secCfg.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to init anylist client: %w", err)
	}

	// fmt.Println("connecting to WS")
	// if err := c.Listen(msgHandler); err != nil {
	// 	return fmt.Errorf("failed to listen for messages: %w", err)
	// }

	var list *List
	refreshList := func(ctx context.Context) error {
		resp, err := c.Lists(ctx)
		if err != nil {
			return fmt.Errorf("failed to load lists: %w", err)
		}
		tmp, err := toList(resp, *groceryListName)
		if err != nil {
			return fmt.Errorf("failed to convert list: %w", err)
		}
		list = tmp
		return nil
	}

	if err := refreshList(ctx); err != nil {
		return fmt.Errorf("failed initial list load: %w", err)
	}

	// This is just a hack to keep the server up to date with other people's
	// changes. We should remove this if we ever get websockets working.
	go func() {
		t := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-t.C:
				refreshList(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(list)
	})
	mux.HandleFunc("/api/add", func(w http.ResponseWriter, r *http.Request) {
		itemName := r.PostFormValue("item_name")
		if err := c.AddItem(r.Context(), list.ID, itemName); err != nil {
			log.Printf("failed to add item %q: %v", itemName, err)
			return
		}
		if err := refreshList(r.Context()); err != nil {
			log.Printf("failed to refresh list: %v", err)
			return
		}
	})
	mux.HandleFunc("/api/remove", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PostFormValue("item_id")
		if err := c.RemoveItem(r.Context(), list.ID, itemID); err != nil {
			log.Printf("failed to remove item %q: %v", itemID, err)
			return
		}
		if err := refreshList(r.Context()); err != nil {
			log.Printf("failed to refresh list: %v", err)
			return
		}
	})
	mux.HandleFunc("/api/check", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.PostFormValue("item_id")
		checked := r.PostFormValue("checked") == "true"
		if err := c.SetChecked(r.Context(), list.ID, itemID, checked); err != nil {
			log.Printf("failed to update checked (%q, %t): %v", itemID, checked, err)
			return
		}
		if err := refreshList(r.Context()); err != nil {
			log.Printf("failed to refresh list: %v", err)
			return
		}
	})
	if err := http.ListenAndServe(":"+strconv.Itoa(*port), cors.Default().Handler(mux)); err != nil {
		return fmt.Errorf("failed to run HTTP server: %w", err)
	}

	return nil
}

type List struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

type Item struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Details string `json:"details"`
	Checked bool   `json:"checked"`
}

func toList(in *pb.PBUserDataResponse, targetListName string) (*List, error) {
	lists := in.ShoppingListsResponse.NewLists
	list, ok := listByName(lists, targetListName)
	if !ok {
		return nil, fmt.Errorf("no list with name %q found", targetListName)
	}

	var items []Item
	for _, item := range list.Items {
		items = append(items, Item{
			ID:      item.Identifier,
			Name:    item.Name,
			Details: item.Details,
			Checked: item.Checked,
		})
	}

	return &List{
		ID:    list.Identifier,
		Name:  list.Name,
		Items: items,
	}, nil
}

func listByName(lists []*pb.ShoppingList, target string) (*pb.ShoppingList, bool) {
	for _, l := range lists {
		if l.Name == target {
			return l, true
		}
	}
	return nil, false
}

func msgHandler(msg string) {
	fmt.Println("Received message!", msg)
}

func decryptConfig(secPath string) (*SecretConfig, error) {
	dat, err := ioutil.ReadFile(secPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret file: %w", err)
	}
	dec, err := decrypt.Data(dat, "json")
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret file: %w", err)
	}
	var sc SecretConfig
	if err := json.Unmarshal(dec, &sc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret config: %w", err)
	}
	return &sc, nil
}
