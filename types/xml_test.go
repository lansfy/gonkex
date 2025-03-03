package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_XML_GetName(t *testing.T) {
	b := &xmlBodyType{}
	require.Equal(t, "XML", b.GetName())
}

func Test_XML_IsSupportedContentType(t *testing.T) {
	b := &xmlBodyType{}
	tests := []struct {
		contentType string
		want        bool
	}{
		{"application/xml", true},
		{"text/xml", true},
		{"application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			require.Equal(t, tt.want, b.IsSupportedContentType(tt.contentType))
		})
	}
}

func Test_XML_Decode(t *testing.T) {
	b := &xmlBodyType{}
	tests := []struct {
		body    string
		want    interface{}
		wantErr string
	}{
		{
			body: "<root><key>value</key></root>",
			want: map[string]interface{}{
				"root": map[string]interface{}{"key": "value"},
			},
		},
		{
			body:    "<invalid_xml>",
			wantErr: "XML syntax error on line 1: unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			got, err := b.Decode(tt.body)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_XML_ExtractResponseValue(t *testing.T) {
	b := &xmlBodyType{}
	tests := []struct {
		body    string
		path    string
		want    string
		wantErr string
	}{
		{
			body: "<root><key>value</key></root>",
			path: "root.key",
			want: "value",
		},
		{
			body:    "<root><key>value</key></root>",
			path:    "missing",
			wantErr: "path '$.missing' does not exist in service response",
		},
		{
			body:    "<invalid_xml>",
			path:    "key",
			wantErr: "invalid XML in response: XML syntax error on line 1: unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.body, func(t *testing.T) {
			got, err := b.ExtractResponseValue(tt.body, tt.path)
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_ParseXML(t *testing.T) {
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
			data, err := ParseXML(tt.rawXml)
			require.NoError(t, err)
			j, err := json.Marshal(data)
			require.NoError(t, err)
			require.JSONEq(t, tt.expectedJson, string(j))
		})
	}
}
