package wallet

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/appditto/pippin_nano_wallet/libs/database"
	"github.com/appditto/pippin_nano_wallet/libs/database/ent"
	"github.com/appditto/pippin_nano_wallet/libs/database/ent/account"
	"github.com/appditto/pippin_nano_wallet/libs/utils"
	"github.com/appditto/pippin_nano_wallet/libs/utils/ed25519"
	"github.com/stretchr/testify/assert"
)

var nanoWallet *NanoWallet
var bananoWallet *NanoWallet

func init() {
	dbconn, _ := database.GetSqlDbConn(true)
	client, _ := database.NewEntClient(dbconn)
	if err := client.Schema.Create(context.TODO()); err != nil {
		panic(err)
	}
	nanoWallet = &NanoWallet{
		DB:     client,
		Ctx:    context.TODO(),
		Banano: false,
	}
	bananoWallet = &NanoWallet{
		DB:     client,
		Ctx:    context.TODO(),
		Banano: true,
	}
}

func TestWalletCreate(t *testing.T) {
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("8d729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040b"))

	wallet, err := nanoWallet.WalletCreate(seed)
	assert.Nil(t, err)

	assert.Equal(t, false, wallet.Encrypted)
	assert.Equal(t, seed, wallet.Seed)

	// Ensure account is created
	account, err := nanoWallet.DB.Account.Query().Where(account.WalletID(wallet.ID)).First(nanoWallet.Ctx)
	assert.Nil(t, err)
	assert.Equal(t, "nano_1efa1gxbitary1urzix9h13nkzadtz71n3auyj7uztb8i4qbtipu8cxz61ee", account.Address)
	assert.Equal(t, 0, account.AccountIndex)
	assert.Equal(t, true, account.Work)
}

func TestWalletCreateBanano(t *testing.T) {
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("bb729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040b"))

	wallet, err := bananoWallet.WalletCreate(seed)
	assert.Nil(t, err)

	assert.Equal(t, false, wallet.Encrypted)
	assert.Equal(t, seed, wallet.Seed)

	// Ensure account is created
	account, err := bananoWallet.DB.Account.Query().Where(account.WalletID(wallet.ID)).First(bananoWallet.Ctx)
	assert.Nil(t, err)
	assert.Equal(t, "ban_33mhuqjxr166czm4y37xk7emfnt4zogxmqrhbfxyngrkbdchmpsk6qehhm3n", account.Address)
	assert.Equal(t, 0, account.AccountIndex)
	assert.Equal(t, true, account.Work)
}

func TestWalletCreateBadInput(t *testing.T) {
	// Empty seed
	_, err := nanoWallet.WalletCreate("")
	assert.ErrorIs(t, ErrInvalidSeed, err)

	// Invalid seed
	_, err = nanoWallet.WalletCreate("invalid seed")
	assert.ErrorIs(t, ErrInvalidSeed, err)
}

func TestAccountCreate(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("5f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040c"))

	wallet, err := nanoWallet.WalletCreate(seed)
	assert.Nil(t, err)

	// Create a couple accounts and ensure they are sequential
	acct, err := nanoWallet.AccountCreate(wallet)
	assert.Nil(t, err)
	assert.Equal(t, 1, acct.AccountIndex)
	assert.Equal(t, "nano_3tdqk8ghsdfapzhrag5978izd19minorfmxergefkecdsbyxaw6og4fejs89", acct.Address)

	acct, err = nanoWallet.AccountCreate(wallet)
	assert.Nil(t, err)
	assert.Equal(t, 2, acct.AccountIndex)
	assert.Equal(t, "nano_1frwge7oebdn87jip7k3sa1uuyf4yxxjh8jg67i69r7smf7tddj1gr6yremf", acct.Address)
}

func TestAccountCreateBadInput(t *testing.T) {
	// Empty seed
	_, err := nanoWallet.AccountCreate(nil)
	assert.ErrorIs(t, ErrInvalidWallet, err)
}

func TestAdhocAccountCreate(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("aa729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040d"))

	wallet, err := nanoWallet.WalletCreate(seed)
	assert.Nil(t, err)

	// Create a couple accounts and ensure they are sequential
	acct, err := nanoWallet.AccountCreate(wallet)
	assert.Nil(t, err)
	assert.Equal(t, 1, acct.AccountIndex)
	assert.Equal(t, "nano_3suihcm3txrfcnipecixeu7kcdm8jisyb1osy14j1r5na1c17g7kkbbu143o", acct.Address)

	_, priv, _ := ed25519.GenerateKey(strings.NewReader("1f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040d"))
	adhocAcct, _, err := nanoWallet.AdhocAccountCreate(wallet, priv)
	assert.Nil(t, err)
	assert.Equal(t, "31663732393334306530376565653639616261633034396332666464346133632c9c941001f4236f487aac98848320ce746f7b809af76e5e795835ef30022424", adhocAcct.PrivateKey)
	assert.Equal(t, "nano_1d6wkia15x35fx69od6rik3k3mmnfxxr38qqfsh9kp3oxwr16b36eztouu3a", adhocAcct.Address)

	// Test duplicate returns same account
	adhocAcct, _, err = nanoWallet.AdhocAccountCreate(wallet, priv)
	assert.Nil(t, err)
	assert.Equal(t, "31663732393334306530376565653639616261633034396332666464346133632c9c941001f4236f487aac98848320ce746f7b809af76e5e795835ef30022424", adhocAcct.PrivateKey)
	assert.Equal(t, "nano_1d6wkia15x35fx69od6rik3k3mmnfxxr38qqfsh9kp3oxwr16b36eztouu3a", adhocAcct.Address)

	// Test we get non-adhoc account if we are trying to re-create it
	_, priv, _ = utils.KeypairFromSeed(seed, 1)
	_, acct, err = nanoWallet.AdhocAccountCreate(wallet, priv)
	assert.Equal(t, "nano_3suihcm3txrfcnipecixeu7kcdm8jisyb1osy14j1r5na1c17g7kkbbu143o", acct.Address)
	assert.Equal(t, 1, acct.AccountIndex)
}

func TestAdhocAccountCreateBadInput(t *testing.T) {
	_, _, err := nanoWallet.AdhocAccountCreate(nil, nil)
	assert.ErrorIs(t, ErrInvalidWallet, err)
	// Bad private key
	_, _, err = nanoWallet.AdhocAccountCreate(&ent.Wallet{}, nil)
	assert.ErrorIs(t, ErrInvalidPrivKey, err)
}

func TestAccountsCreate(t *testing.T) {
	os.Setenv("MOCK_REDIS", "true")
	defer os.Unsetenv("MOCK_REDIS")
	// Predictable seed
	seed, _ := utils.GenerateSeed(strings.NewReader("9f729340e07eee69abac049c2fdd4a3c4b50e4672a2fabdf1ae295f2b4f3040e"))

	wallet, err := nanoWallet.WalletCreate(seed)
	assert.Nil(t, err)

	// Create a couple accounts and ensure they are sequential
	accts, err := nanoWallet.AccountsCreate(wallet, 10)
	assert.Nil(t, err)
	assert.Len(t, accts, 10)
	assert.Equal(t, 1, accts[0].AccountIndex)
	assert.Equal(t, 2, accts[1].AccountIndex)
	assert.Equal(t, 3, accts[2].AccountIndex)
	assert.Equal(t, 4, accts[3].AccountIndex)
	assert.Equal(t, 5, accts[4].AccountIndex)
	assert.Equal(t, 6, accts[5].AccountIndex)
	assert.Equal(t, 7, accts[6].AccountIndex)
	assert.Equal(t, 8, accts[7].AccountIndex)
	assert.Equal(t, 9, accts[8].AccountIndex)
	assert.Equal(t, 10, accts[9].AccountIndex)
	assert.Equal(t, "nano_3hntkbk1q6pn1n8481shemojcmtpxxpjbojm7h5h5p6jz53bahjuif6d8j4f", accts[0].Address)
	assert.Equal(t, "nano_1dh7j8bw1pi8xur1zste3c7oqc1ef3yttory4snhufb5haj1ctzxapgpqhzq", accts[1].Address)
	assert.Equal(t, "nano_1qhc3k9p1dxb45f9jrsz3afydqdxoo8y4deeu75q34pmskjgyafq67fd5k1r", accts[2].Address)
	assert.Equal(t, "nano_3cmpq338qbq6hhhijia6ack71kas64snza6r89zhspffzaimxnokaofpa1wt", accts[3].Address)
	assert.Equal(t, "nano_319uakdyoq48h3zswtwt8ei7fjdq4mue3dizzz3fbgnteqb5zqeycygteerd", accts[4].Address)
	assert.Equal(t, "nano_1xew8dmxca5fbo16hgg88p9sgfpamig1kdos5te68hkstj3kf4bi1cyb3qdt", accts[5].Address)
	assert.Equal(t, "nano_33iiz1hkggxirkspip8hef9uiiguhay5ykjdydnyuresxxt1mkkrfz9qeiak", accts[6].Address)
	assert.Equal(t, "nano_3gxjsq4b55fd6hmzroa7wgtmw7wy7nwjofspewzhdwdo5dck9sa85rnfujrn", accts[7].Address)
	assert.Equal(t, "nano_3xoyiewop6gk4g8jaynkf7bhn5baciwtdppkkoprsbafmimcmaiywz6pjqt5", accts[8].Address)
	assert.Equal(t, "nano_1pigk4bbfhpdg3u5b8sdhij8c78n66eap9478p4dhgamyqt7herco57z6ib7", accts[9].Address)
}

func TestAccountsCreateBadInput(t *testing.T) {
	_, err := nanoWallet.AccountsCreate(nil, 10)
	assert.ErrorIs(t, ErrInvalidWallet, err)
	_, err = nanoWallet.AccountsCreate(&ent.Wallet{}, 0)
	assert.ErrorIs(t, ErrInvalidAccountCount, err)
}