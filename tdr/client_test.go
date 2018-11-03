package tdr

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/philips-software/go-hsdp-api/iam"
)

var (
	muxIAM    *http.ServeMux
	serverIAM *httptest.Server
	muxIDM    *http.ServeMux
	serverIDM *httptest.Server
	muxTDR    *http.ServeMux
	serverTDR *httptest.Server

	iamClient *iam.Client
	tdrClient *Client
)

func setup(t *testing.T) func() {
	muxIAM = http.NewServeMux()
	serverIAM = httptest.NewServer(muxIAM)
	muxIDM = http.NewServeMux()
	serverIDM = httptest.NewServer(muxIDM)
	muxTDR = http.NewServeMux()
	serverTDR = httptest.NewServer(muxTDR)

	var err error
	iamClient, err = iam.NewClient(nil, &iam.Config{
		OAuth2ClientID: "TestClient",
		OAuth2Secret:   "Secret",
		SharedKey:      "SharedKey",
		SecretKey:      "SecretKey",
		IAMURL:         serverIAM.URL,
		IDMURL:         serverIDM.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create iamCleitn: %v", err)
	}
	token := "44d20214-7879-4e35-923d-f9d4e01c9746"

	muxIAM.HandleFunc("/authorize/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected ‘POST’ request")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{
    "scope": "mail tdr.contract tdr.dataitem",
    "access_token": "`+token+`",
    "refresh_token": "31f1a449-ef8e-4bfc-a227-4f2353fde547",
    "expires_in": 1799,
    "token_type": "Bearer"
}`)
	})

	// Login immediately so we can create tdrClient
	iamClient.Login("username", "password")

	tdrClient, err = NewClient(nil, iamClient, &Config{
		TDRURL: serverTDR.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create tdrClient: %v", err)
	}

	return func() {
		serverIAM.Close()
		serverIDM.Close()
	}
}

func TestLogin(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	token := "44d20214-7879-4e35-923d-f9d4e01c9746"

	err := iamClient.Login("username", "password")
	if err != nil {
		t.Fatal(err)
	}
	if iamClient.Token() != token {
		t.Errorf("Unexpected token found: %s, expected: %s", iamClient.Token(), token)
	}
	if !iamClient.HasScopes("tdr.contract", "tdr.dataitem") {
		t.Errorf("Client should have tdr.contract and tdr.dataitem scopes")
	}
}

func TestDebug(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	tdrClient, err = NewClient(nil, iamClient, &Config{
		TDRURL:   serverTDR.URL,
		Debug:    true,
		DebugLog: tmpfile.Name(),
	})
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	defer tdrClient.Close()
	defer os.Remove(tmpfile.Name()) // clean up

	iamClient.Login("username", "password")
	tdrClient.Contracts.GetContract(&GetContractOptions{
		Datatype: String("TestGo|TestGoContract"),
	}, nil)

	fi, err := tmpfile.Stat()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if fi.Size() == 0 {
		t.Errorf("Expected something to be written to DebugLog")
	}

}
