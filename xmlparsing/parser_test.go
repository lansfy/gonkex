package xmlparsing

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		description  string
		rawXml       string
		expectedJson string
	}{
		{
			description: "small xml",
			rawXml: `
		<?xml version="1.0" encoding="UTF-8"?>
		<Person>
			<Company><![CDATA[Hogwarts School of Witchcraft and Wizardry]]></Company>
			<FullName>Harry Potter</FullName>
			<Email where="work">hpotter@hog.gb</Email>
			<Email where="home">hpotter@gmail.com</Email>
			<Addr>4 Privet Drive</Addr>
			<Group>
				<Value>Hexes</Value>
				<Value>Jinxes</Value>
			</Group>
		</Person>`,
			expectedJson: `
			{
				"Person": {
					"Company": "Hogwarts School of Witchcraft and Wizardry",
					"FullName": "Harry Potter",
					"Email": [
					{
						"-attrs": {"where": "work"},
						"content": "hpotter@hog.gb"
					},
					{
						"-attrs": {"where": "home"},
						"content": "hpotter@gmail.com"
					}
					],
					"Addr": "4 Privet Drive",
					"Group": {
						"Value": ["Hexes", "Jinxes"]
					}
				}
			}`,
		},
		{
			description: "namespaces",
			rawXml: `
		<ns1:person>
			<ns2:name>Eddie</ns2:name>
			<ns2:surname>Dean</ns2:surname>
		</ns1:person>
		`,
			expectedJson: `
			{
				"ns1:person": {
					"ns2:name": "Eddie",
					"ns2:surname": "Dean"
				}
			}`,
		},
		{
			description: "empty tag",
			rawXml:      "<body><emptytag/></body>",
			expectedJson: `
			{
				"body": {
					"emptytag": ""
				}
			}`,
		},
		{
			description: "only attributes",
			rawXml:      `<body><tag attr1="attr1_value" attr2="attr2_value"/></body>`,
			expectedJson: `
			{
				"body": {
					"tag": {
						"-attrs": {
							"attr1": "attr1_value",
							"attr2": "attr2_value"
						},
						"content": ""
					}
				}
			}`,
		},
		{
			description: "xmd document #2",
			rawXml: `
		<?xml version="1.0" encoding="UTF-8"?>
		<Items>
			<Item>
				<Name>name1</Name>
				<Value>value1</Value>
			</Item>
			<Item>
				<Name>name2</Name>
				<Value>value2</Value>
			</Item>
		</Items>`,
			expectedJson: `
			{
				"Items": {
					"Item": [
						{"Name":"name1","Value":"value1"},
						{"Name":"name2","Value":"value2"}
					]
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			data, err := Parse(tt.rawXml)
			require.NoError(t, err)
			j, err := json.Marshal(data)
			require.NoError(t, err)
			require.JSONEq(t, tt.expectedJson, string(j))
		})
	}
}
