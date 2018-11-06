package tdr

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetContract(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	muxTDR.HandleFunc("/store/tdr/Contract", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("dataType") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{
			"type": "searchset",
			"total": 0,
			"entry": [],
			"resourceType": "Bundle"
		  }`)
	})
	contracts, resp, err := tdrClient.Contracts.GetContract(&GetContractOptions{
		Datatype: String("TestGo|TestGoContract"),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP success Got: %d", resp.StatusCode)
	}
	if len(contracts) != 0 {
		t.Errorf("Expected 0 contracts for now")
	}
}

func TestCreateContract(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	muxTDR.HandleFunc("/store/tdr/Contract", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Unexpected EOF from reading request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var contract Contract
		err = json.Unmarshal(body, &contract)
		if err != nil {
			t.Errorf("Expected contract in body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/need/to/capture/this")
		w.WriteHeader(http.StatusCreated)
	})

	var schemaContract = []byte(`{
		"$schema": "http://json-schema.org/draft-04/schema#",
		"type": "object",
		"properties": {
		  "Temperature": {
			"type": "number"
		  },
		  "HeartRate": {
			"type": "integer"
		  },
		  "IsManualMeasurement": {
			"type": "boolean"
		  },
		  "DeviceStatus": {
			"type": "string"
		  }
		},
		"required": [
		  "Temperature",
		  "HeartRate"
		]
	  }`)

	var newContract = Contract{
		SendNotifications: false,
		Organization:      "DevOrg",
		DataType: DataType{
			System: "TestGo",
			Code:   "TestGoContract",
		},
		DeletePolicy: DeletePolicy{
			Duration: 1,
			Unit:     "MONTH",
		},
		Schema: json.RawMessage(schemaContract),
	}

	ok, resp, err := tdrClient.Contracts.CreateContract(newContract)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected HTTP created Got: %d", resp.StatusCode)
	}
	if !ok {
		t.Errorf("Contract creation failed")
	}
}