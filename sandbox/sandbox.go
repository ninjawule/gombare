package main

import "encoding/json"

func main() {

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
	
								"when": [{"prop": "type", "is": "BVM", 
									"look": [{"composant": {"when": [{"prop": "@organe", "is": "BV", 
										"look": [{"propriete": {"when": [{"prop": "@operation", "is": "TBOIT",
											"_use": ["#text"]}]}}]}]}}]}],
	
								"look":[
									{".": {"_use": ["@type"]}},
									{"date": {"_use": ["datedebut", "datefin"]}},
									{"critere": {"_use": ["datedebut", "datefin"]}}],
	
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

	param2 := &IdentificationParameter{}
	if err3 := json.Unmarshal(dataBytes, param2); err3 != nil {
		panic(err3)
	}

	dataBytes2, err4 := json.MarshalIndent(param2, "", "	")
	if err4 != nil {
		panic(err4)
	}

	//nolint
	println(string(dataBytes2))
}

type IdentificationParameter struct {
	Path string                                `json:"path,omitempty"`
	Use  []string                              `json:"_use,omitempty"`
	When []*ConditionalIDParameter             `json:"when,omitempty"`
	Look []map[string]*IdentificationParameter `json:"look,omitempty"` // array of 1 pair of key-value
	For  map[string]*IdentificationParameter   `json:"_for,omitempty"`
}

type ConditionalIDParameter struct {
	Prop string `json:"prop,omitempty"`
	Is   string `json:"is,omitempty"`
	IdentificationParameter
}
