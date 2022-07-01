package main

import "encoding/json"

func main() {
	//nolint
	data := `{
		"_for": {
			"data": {
				"_for": {
					"vehicule": {
						"_use": ["vmoclecode"],
						"_for": {
							"carrosserie": {
								"look": [
									{"volume": {"_use": ["#text"]}}]
							},
							
							"ensemble": {
	
								"when": [
									{"prop": "@type","is": "BVM",
										"look": [{"at": "composant", "when": [{"prop": "@organe", "is": "BV",
											"look": [{"at": "propriete", "when": [{"prop": "@operation","is": "TBOIT",
												"_use": ["#text"]}]}]}]}]}
								],
	
								"look":[
									{"at":".", "_use": ["@type"]},
									{"at":"date", "_use": ["datedebut", "datefin"]},
									{"at":"critere", "_use": ["datedebut", "datefin"]}],
	
								"_for": {
									"composant": {
										"_use": ["@organe"],
										"_for": {
											"propriete": {
												"_use": ["@operation"]
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	param := &IdentificationParameter{}

	if err := json.Unmarshal([]byte(data), param); err != nil {
		panic(err)
	}

	dataBytes, err2 := json.MarshalIndent(param, "", "	")
	if err2 != nil {
		panic(err2)
	}

	//nolint
	println(string(dataBytes))
}

type IdentificationParameter struct {
	At   string                              `json:"at,omitempty"`
	Use  []string                            `json:"_use,omitempty"`
	When []*ConditionalIDParameter           `json:"when,omitempty"`
	Look []*IdentificationParameter          `json:"look,omitempty"` // array of 1 pair of key-value
	For  map[string]*IdentificationParameter `json:"_for,omitempty"`
}

type ConditionalIDParameter struct {
	Prop string `json:"prop,omitempty"`
	Is   string `json:"is,omitempty"`
	IdentificationParameter
}
