# Gonkex: testing automation tool

Gonkex will test your services using their API. It can bomb the service with prepared requests and check the responses. Test scenarios are described in YAML-files.

Capabilities:

- works with REST/JSON API
- seeds the DB with fixtures data (supports PostgreSQL, MySQL, Sqlite, TimescaleDB, MariaDB, SQLServer, ClickHouse)
- provides mocks for external services
- can be used as a library and ran together with unit-tests
- stores the results as an [Allure](https://allurereport.org/) report
- there is a [JSON-schema](#json-schema) to add autocomletion and validation for gonkex YAML files

## Table of contents

- [Using gonkex as a library](#using-gonkex-as-a-library)
- [Test scenario example](#test-scenario-example)
- [Test status](#test-status)
- [HTTP-request](#http-request)
- [HTTP-response](#http-response)
- [Variables](#variables)
  - [Assignment](#assignment)
    - [In the description of the test](#in-the-description-of-the-test)
    - [From the response of the previous test](#from-the-response-of-the-previous-test)
    - [From the response of currently running test](#from-the-response-of-currently-running-test)
    - [From environment variables or from env-file](#from-environment-variables-or-from-env-file)
    - [From cases](#from-cases)
- [Files uploading](#files-uploading)
- [Fixtures](#fixtures)
  - [Deleting data from tables](#deleting-data-from-tables)
  - [Record templates](#record-templates)
  - [Record inheritance](#record-inheritance)
  - [Expressions](#expressions)
- [Mocks](#mocks)
  - [Running mocks while using gonkex as a library](#running-mocks-while-using-gonkex-as-a-library)
  - [Mocks definition in the test file](#mocks-definition-in-the-test-file)
    - [Request constraints (requestConstraints)](#request-constraints-requestconstraints)
    - [Response strategies (strategy)](#response-strategies-strategy)
    - [Calls count](#calls-count)
- [Shell scripts usage](#shell-scripts-usage)
  - [Script definition](#script-definition)
  - [Running a script with parameterization](#running-a-script-with-parameterization)
- [A DB query](#a-db-query)
  - [Test Format](#test-format)
  - [Query definition](#query-definition)
  - [Definition of DB request response](#definition-of-db-request-response)
  - [DB request parameterization](#db-request-parameterization)
  - [Ignoring ordering in DB response](#ignoring-ordering-in-db-response)
- [JSON-schema](#json-schema)
  - [Setup in Jetbrains IDE](#setup-in-jetbrains-ide)
  - [Setup is VSCode IDE](#setup-is-vscode-ide)

## Using gonkex as a library

To integrate functional and native Go tests and run them together, use gonkex as a library.

Create a test file, for example `func_test.go`.

Import gonkex as a dependency to your project in this file:

```go
import (
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/runner"
)
```

Create a test function:

```go
package test

import (
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/storage"
	"github.com/lansfy/gonkex/storage/addons/sqldb"
	"github.com/lansfy/gonkex/runner"
)

func TestFuncCases(t *testing.T) {
	// init the mocks if needed (details below)
	// m := mocks.NewNop(...)

	// init the DB to load the fixtures if needed (details below)
	//
	// db := ...
	// storage := sqldb.NewStorage(sqldb.PostgreSQL, db, nil)
	//
	// next sql storages supported:
	//    sqldb.PostgreSQL,  sqldb.MySQL,   sqldb.Sqlite,  sqldb.ClickHouse,
	//    sqldb.TimescaleDB, sqldb.MariaDB, sqldb.SQLServer

	// create a server instance of your app
	srv := server.NewServer()
	defer srv.Close()

	// run test cases from current folder
	runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingParams{
		TestsDir:    "cases",      // test case folder
		FixturesDir: "fixtures",   // fixtures folder
		Mocks:       m,
		DB:          storage,
	})
}
```

Externally written storage may be used for loading test data, if gonkex used as a library.
To start using the custom storage, you need to import the custom module, that contains implementation of storage.StorageInterface interface.
For example, the following NoSQL databases are currently supported as custom modules:
- Aerospike (github.com/lansfy/gonkex/storage/addons/aerospike)
- MongoDB (github.com/lansfy/gonkex/storage/addons/mongo)
- Redis (github.com/lansfy/gonkex/storage/addons/redis)

The tests can be now ran with `go test`, for example: `go test ./...`.

## Test scenario example

```yaml
- name: WHEN the list of orders is requested service MUST return selected order
  method: GET
  status: ""
  path: /jsonrpc/v2/order.getBriefList
  query: ?id=11111111-1111-1111-1111-aaaaaaaaaaaa&jsonrpc=2.0&user_id=00001

  fixtures:
    - order_0001
    - order_0002

  response:
    200: |
      {
        "id": "11111111-1111-1111-1111-aaaaaaaaaaaa",
        "jsonrpc": "2.0",
        "result": {
          "data": [
            "ORDER0001",
            "ORDER0002"
          ],
          "meta": {
            "items": 0,
            "limit": 50,
            "page": 0,
            "pages": 0
          }
        }
      }

- name: WHEN one order is requested service MUST response with user and order sum
  method: POST
  path: /jsonrpc/v2/order.getOrder

  headers:
    Authorization: Bearer HsHG67d38hJKJFdfjj==
    Content-Type: application/json

  cookies:
    sid: ZmEwZDkwYzgwMmQzMGIzOGIxODM3ZmFiOTGJhMzU=
    lid: AAAEAFu/TdhHBg7UAgA=

  request: |
    {
      "jsonrpc": "2.0",
      "id": "11111111-1111-1111-1111-aaaaaaaaaaaa",
      "method": "order.getOrder",
      "params": [
        {
          "order_nr": {{ .orderNr }}
        }
      ]
    }

  comparisonParams:
    ignoreValues: false
    ignoreArraysOrdering: false
    disallowExtraFields: false

  response:
    200: |
      {
        "id": "11111111-1111-1111-1111-aaaaaaaaaaaa",
        "jsonrpc": "2.0",
        "result": {
          "user_id": {{ .userId }},
          "amount": {{ .amount }},
          "token": "$matchRegexp(^\\w{16}$)"
        }
      }

  responseHeaders:
    200:
      Content-Type: "application/json"
      Cache-Control: "no-store, must-revalidate"
      Set-Cookie: "mycookie=123; Path=/; Domain=mydomain.com", "mycookie=456; Path=/; Domain=.mydomain.com"

  cases:
    - requestArgs:
        orderNr: ORDER0001
      responseArgs:
        200:
          userId: '0001'
          amount: 1000

    - requestArgs:
        orderNr: ORDER0002
      responseArgs:
        200:
          userId: '0001'
          amount: 72000
```

Prefix "?" in query field is optional.

As you can see in this example, you can use Regexp for checking response body.
It can be used for whole body (if it's plain text):

```yaml
    response:
        200: "$matchRegexp(^xy+z$)"
```

or for elements of map/array (if it's JSON):

```yaml
  response:
    200: >
      {
        "id": "$matchRegexp([\\w-]+)",
        "jsonrpc": "$matchRegexp([12].0)",
        "result": [
          "data": [
              "$matchRegexp(^ORDER[0]{3}[0-9]$)",
              "$matchRegexp(^ORDER[0]{3}[0-9]$)"
          ]
        ]
      }
```

## HTTP-request

`method` - a parameter for HTTP request type (e.g. "GET", "POST", "DELETE" and so on)

`path` - a parameter for URL path, the format is in the example above.

`headers` - a parameter for HTTP headers, the format is in the example above.

`cookies` - a parameter for cookies, the format is in the example above.

## HTTP-response

`response` - the HTTP response body for the specified HTTP status codes.

`responseHeaders` - all HTTP response headers for the specified HTTP status codes.

## Test status

`status` - a parameter, for specially mark tests, can have following values:

- `broken` - do not run test, only mark it as broken
- `skipped` - do not run test, only mark it as skipped
- `focus` - run only this specific test, and mark all other tests with unset status as `skipped`

## Retry policy

If you expect a test to succeed after only a few attempts (for example, one testcase has run some asynchronous operation and the second testcase is trying to wait for the results after that),
then you need to do several test retry. You can define the number of retries required using the `retryPolicy` field.

Example:

```yaml
 - name: wait for operation result
   method: GET
   ...
   retryPolicy:
     attempts: 6         # retry failed test 6 times
     delay: 5s           # with 5 second delay between retries
     successInRow: 2     # it takes 2 successful test runs to recognize the test as successful
```

Next fields supported:

`attempts` - an integer indicating the number of times that gonkex will retry the test request in the event assertions fail.

`delay` - string containing the waiting time after unsuccessful completion of the test.

`successInRow` - parameter defines the required number of successful test passes for the test to be recognized as successful. And all these successful runs must be consecutive. Default value is 1.

## Variables

You can use variables in the description of the test, the following fields are supported:

- method
- description
- path
- query
- headers
- request
- response
- dbQuery
- dbResponse
- mocks body
- mocks headers
- mocks requestConstraints
- form for multipart/form-data

Example:

```yaml
- method: "{{ $method }}"
  description: "{{ $description }}"
  path: "/some/path/{{ $pathPart }}"
  query: "{{ $query }}"
  headers:
    header1: "{{ $header }}"
  request: '{"reqParam": "{{ $reqParam }}"}'
  response:
    200: "{{ $resp }}"
  mocks:
    server_mock:
      strategy: constant
      body: >
        {
          "message": "{{ $mockParam }}"
        }
      statusCode: 200
  dbChecks:
    - dbQuery: "SELECT id, name FROM testing_tools WHERE id={{ $sqlQueryParam }}"
      dbResponse:
        - '{"id": {{ $sqlResultParam }}, "name": "test"}'
```

You can assign values to variables in the following ways (priorities are from top to bottom):

- in the description of the test
- from the response of the previous test
- from the response of currently running test
- from environment variables or from env-file

### Assignment

#### In the description of the test

Example:

```yaml
- method: "{{ $someVar }}"
  path: "/some/path/{{ $someVar }}"
  query: "{{ $someVar }}"
  headers:
    header1: "{{ $someVar }}"
  request: '{"reqParam": "{{ $someVar }}"}'
  response:
    200: "{{ $someVar }}"
  variables:
    someVar: "someValue"
```

#### From the response of the previous test

Example:

```yaml
# if the response is plain text
- name: "get_last_post_id"
  ...
  variables_to_set:
    200: "id"

# if the response is JSON
- name: "get_last_post_info"
  ...
  variables_to_set:
    200:
      id: "id"
      title: "title"
      authorId: "author_info.id"
```

You can access nested json fields like this:
> "author_info.id"

Any nesting levels are supported.

#### From the response of currently running test

Example:

```yaml
- name: Get info with database
  method: GET
  path: /info/1
  variables_to_set:
    200:
      golang_id: "query_result.0.0"
  response:
    200: '{"result_id": "1", "query_result": [[ {{ $golang_id }}, "golang"], [2, "gonkex"]]}'
  dbChecks:
    - dbQuery: "SELECT id, name FROM testing_tools WHERE id={{ $golang_id }}"
      dbResponse:
        - '{"id": {{ $golang_id}}, "name": "golang"}'
```

#### From environment variables or from env-file

Gonkex automatically checks if variable exists in the environment variables (case-sensitive) and loads a value from there, if it exists.

If an env-file is specified, variables described in it will be added or will replace the corresponding environment variables.

Example of an env file (standard syntax):

```.env
jwt=some_jwt_value
secret=my_secret
password=private_password
```

env-file can be convenient to hide sensitive information from a test (passwords, keys, etc.) or specify common used values here.

#### From cases

You can describe variables in *cases* section of a test.

Example:

```yaml
- name: Get user info
  method: GET
  path: /user/1
  response:
    200: '{ "user_id": "1", "name": "{{ $name }}", "surname": "{{ $surname }}" }'
  cases:
    - variables:
        name: John
        surname: Doe
```

Variables like these will be available through another cases if not redefined.

## multipart/form-data requests
You must specify the type of request:
- POST

Header (optional):
> Content-Type: multipart/form-data

with _boundary_ (optional):
> Content-Type: multipart/form-data; boundary=--some-boundary


### Form

Example:

```yaml
 - name: "upload-form"
   method: POST
   form:
     fields:
       field_name1: "field_name1 value"
       field_name2: "field_name2 value"
       "custom_struct_field[0]": "custom_struct_field 0"
       "custom_struct_field[1]": "custom_struct_field 1"
       "custom_struct_field[inner_obj][field]": "inner_obj field value"
   headers:
     Content-Type: multipart/form-data # case-sensitive, can be omitted
   response:
     200: |
       {
         "status": "OK"
       }
```

### File upload

You can upload files in test request.
Example:

```yaml
 - name: "upload-files"
   method: POST
   form:
     files:
       file1: "testdata/upload-files/file1.txt"
       file2: "testdata/upload-files/file2.log"
   headers:
     Content-Type: multipart/form-data
   response:
     200: >
       {
         "status": "OK"
       }
```

with form:

```yaml
 - name: "upload-multipart-form-data"
   method: POST
   form:
     fields:
       field_name1: "field_name1 value"
     files:
       file1: "testdata/upload-files/file1.txt"
       file2: "testdata/upload-files/file2.log"
   headers:
     Content-Type: multipart/form-data
   response:
     200: >
       {
         "status": "OK"
       }
```

## Fixtures

To seed the DB before the test, gonkex uses fixture files.

File example:

```yaml
# fixtures/comments.yml
inherits:
  - another_fixture
  - yet_another_fixture

tables:
  posts:
    - $name: janes_post
      title: New post
      text: Post text
      author: Jane Dow
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - $name: apples_post
      title: Morning digest
      text: Text
      author: Apple Seed
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  comments:
    - post_id: $janes_post.id
      content: A comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - post_id: $apples_post.id
      content: Another comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  another_table:
    ...
  ...
```

Records in fixtures can use templates, inherit and reference each other.

### Deleting data from tables

To clear the table before the test put square brackets next to the table name.

Example:

```yaml
# fixtures/empty_posts_table.yml
tables:
  posts: []
```

### Record templates

Usually, to insert a record to a DB, it's necessary to list all the fields without default values.
Oftentimes, many of those fields are not important for the test, and their values repeat from one fixture to another, creating unnecessary visual garbage and making the maintenance harder.

With templates you can inherit the fields from template record redefining only the fields that are important for the test.

Template definition example:

```yaml
templates:
  dummy_client:
    name: Dummy Client Name
    age: 35
    ip: 127.0.0.1
    is_deleted: false

  dummy_deleted_client:
    $extend: dummy_client
    is_deleted: true

tables:
  ...
```

Example of using a template in a fixture:

```yaml
templates:
  ...
tables:
  clients:
    - $extend: dummy_client
    - $extend: dummy_client
      name: Josh
    - $extend: dummy_deleted_client
      name: Jane
```

As you might have noticed, templates can be inherited as well with `$extend` keyword, but only if by the time of the dependent template definition the parent template is already defined (in this file or any other referenced with `inherits`).

### Record inheritance

Records can be inherited as well using `$extend`.

To inherit a record, first you need to assign this record a name using `$name`:

```yaml
# fixtures/post.yaml
tables:
  posts:
    - $name: regular_post
      title: Post title
      text: Some text
```

Names assigned to records must be unique among all loaded fixture files, as well as they must not interfere with template names.

In another fixture file you need to declare that a certain record inherits an earlier defined record with `$extend`, just like with the templates:

```yaml
# fixtures/deleted_post.yaml
inherits:
  - post
tables:
  posts:
    - $extend: regular_post
      is_deleted: true
```

Don't forget to declare the dependency between files in `inherits`, to make sure that one file is always loaded together with the other one.

It's important to note that record inheritance only works with different fixture files.
It's not possible to declare inheritance within one file.

### Expressions

When you need to write an expression execution result to the DB and not a static value, you can use `$eval()` construct.
Everything inside the brackets will be inserted into the DB as raw, non-escaped data.
This way, within `$eval()` you can write everything you would in a regular query.

For instance, this construct allows the insertion of current date and time as a field value:

```yaml
tables:
  comments:
    - created_at: $eval(NOW())
```
## Mocks

In order to imitate responses from external services, use mocks.

A mock is a web server that is running on-the-fly, and is populated with certain logic before the execution of each test.
The logic defines what the server responses to a certain request. It's defined in the test file.

### Running mocks while using gonkex as a library

Before running tests, all planned mocks are started.
It means that gonkex spins up the given number of servers and each one of them gets a random port assigned.

```go
// create empty server mocks
m := mocks.NewNop(
	"cart",
	"catalog",
	"loyalty",
	"discounts",
)

// spin up mocks
err := m.Start()
if err != nil {
	t.Fatal(err)
}
defer m.Shutdown()
```

After spinning up the mock web-servers, we can get their addresses (host and port).
Using those addresses, you can configure your service to send their requests to mocked servers instead of real ones.

```go
// configuring and running the service
srv := server.NewServer(&server.Config{
	CartAddr:      m.Service("cart").ServerAddr(),
	CatalogAddr:   m.Service("catalog").ServerAddr(),
	LoyaltyAddr:   m.Service("loyalty").ServerAddr(),
	DiscountsAddr: m.Service("discounts").ServerAddr(),
})
defer srv.Close()
```

Additionally, library registers special environment variables `GONKEX_MOCK_<MOCK_NAME>` the for every mock, which contain the address and port of the corresponding mock server.
You can use these environment variables when writing tests.

As soon as you spinned up your mocks and configured your service, you can run the tests.

```go
runner.RunWithTesting(t, srv.URL, &runner.RunWithTestingParams{
	TestsDir: "tests/cases",
	Mocks:    m, // pass the mocks to the test runner
})
```

### Mocks definition in the test file

Each test communicates a configuration to the mock-server before running. This configuration defines the responses for specific requests in the mock-server.
The configuration is defined in a YAML-file with test in the `mocks` section.

The test file can contain any number of mock service definitions:

```yaml
- name: Test with mocks
  request:
    ...
  ...
  mocks:
    service1:
      ...
    service2:
      ...
    service3:
      ...
```

Each mock-service definition consists of:

`requestConstraints` - an array of constraints that are applied on a received request. If at least one constraint is not satisfied, the test is considered failed. The list of all possible checks is provided below.

`strategy` - the strategy of mock responses. The list of all possible strategies is provided below.

The rest of the keys on the first nesting level are parameters to the strategy. Their variety is different for each strategy.

A configuration example for one mock-service:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - ...
        - ...
      strategy: strategyName
      strategyParam1: ...
      strategyParam2: ...
    ...
```

#### Request constraints (requestConstraints)

The request to the mock-service can be validated using one or more constraints defined below.

The definition of each constraint contains of the `kind` parameter that indicates which constraint will be applied.

All other keys on this level are constraint parameters. Each constraint has its own parameter set.

##### nop

Empty constraint. Always successful.

No parameters.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: nop
    ...
```

##### bodyMatchesJSON

Checks that the request body is JSON, and it corresponds to the JSON defined in the `body` parameter.

Parameters:

- `body` (mandatory) - expected JSON. All keys on all levels defined in this parameter must be present in the request body.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # this check will demand that the request contains keys key1, key2 and subKey1
        # and their values set to value1 and value2. However, it's fine if the request has
        # other keys not mentioned here.
        - kind: bodyMatchesJSON
          body: >
            {
              "key1": "value1",
              "key2": {
                "subKey1": "value2",
              }
            }
    ...
```

##### bodyJSONFieldMatchesJSON

When request body is JSON, checks that value of particular JSON-field is string-packed JSON
that matches to JSON defined in `value` parameter.

Parameters:

- `path` (mandatory) - path to string field, containing JSON to check.
- `value` (mandatory) - expected JSON.

Example:

Origin request that contains string-packed JSON

```yaml
  {
      "field1": {
        "field2": "{\"stringpacked\": \"json\"}"
      }
  }
```

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyJSONFieldMatchesJSON
          path: field1.field2
          value: |
            {
              "stringpacked": "json"
            }
  ...
```

##### pathMatches

Checks that the request path corresponds to the expected one.

Parameters:

- `path` - a string with the expected request path value;
- `regexp` - a regular expression to check the path value against.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: pathMatches
          path: /api/v1/test/somevalue
    service2:
      requestConstraints:
        - kind: pathMatches
          regexp: ^/api/v1/test/.*$
    ...
```

##### queryMatches

Checks that the GET request parameters correspond to the ones defined in the `query` parameter.

Parameters:

- `expectedQuery` (mandatory) - a list of parameters to compare the parameter string to. The order of parameters is not important.

Examples:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # this check will demand that the request contains key1 и key2
        # and the values are key1=value1, key1=value11 и key2=value2.
        # Keys not mentioned here are omitted while running the check.
        - kind: queryMatches
          expectedQuery:  key1=value1&key2=value2&key1=value11
    ...
```

##### queryMatchesRegexp

Expands `queryMatches` so it can be used with regexp pattern matching.

Parameters:

- `expectedQuery` (mandatory) - a list of parameters to compare the parameter string to. The order of parameters is not important.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # works similarly to queryMatches with an addition of $matchRegexp usage
        - kind: queryMatchesRegexp
          expectedQuery:  key1=value1&key2=$matchRegexp(\\d+)&key1=value11
    ...
```

##### methodIs

Checks that the request method corresponds to the expected one.

Parameters:

- `method` (mandatory) - string to compare the request method to.

There are also 2 short variations that don't require `method` parameter:

- `methodIsGET`
- `methodIsPOST`

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: methodIs
          method: PUT
    service2:
      requestConstraints:
        - kind: methodIsPOST
    ...
```

##### headerIs

Checks that the request has the defined header and (optional) that its value either equals the pre-defined one or falls under the definition of a regular expression.

Parameters:

- `header` (mandatory) - name of the header that is expected with the request;
- `value` - a string with the expected request header value;
- `regexp` - a regular expression to check the header value against.

Examples:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          value: application/json
    service2:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          regexp: ^(application/json|text/plain)$
    ...
```

##### bodyMatchesText

Checks that the request has the defined body text, or it falls under the definition of a regular expression.

Parameters:

- `body` - a string with the expected request body value;
- `regexp` - a regular expression to check the body value against.

Examples:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesText
          body: |-
            query HeroNameAndFriends {
                  hero {
                    name
                    friends {
                      name
                    }
                  }
                }
    service2:
      requestConstraints:
        - kind: bodyMatchesText
          regexp: (HeroNameAndFriends)
    ...
```

##### bodyMatchesXML

Checks that the request body is XML, and it matches to the XML defined in the `body` parameter.

Parameters:

- `body` (mandatory) - expected XML.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesXML
          body: |
            <Person>
              <Company>Hogwarts School of Witchcraft and Wizardry</Company>
              <FullName>Harry Potter</FullName>
              <Email where="work">hpotter@hog.gb</Email>
              <Email where="home">hpotter@gmail.com</Email>
              <Addr>4 Privet Drive</Addr>
              <Group>
                <Value>Hexes</Value>
                <Value>Jinxes</Value>
              </Group>
            </Person>
  ...
```

#### Response strategies (strategy)

Response strategies define what mock will response to incoming requests.

##### nop

Empty strategy. All requests are served with `204 No Content` and empty body.

No parameters.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: nop
    ...
```

##### file

Returns a response read from a file.

Parameters:

- `filename` (mandatory) - name of the file that contains the response body;
- `statusCode` - HTTP-code of the response, the default value is `200`;
- `headers` - response headers.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: file
      filename: responses/service1_success.json
      statusCode: 500
      headers:
        Content-Type: application/json
    ...
```

##### constant

Returns a defined response.

Parameters:

- `body` (mandatory) - sets the response body;
- `statusCode` - HTTP-code of the response, the default value is `200`;
- `headers` - response headers.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: constant
      body: >
        {
          "status": "error",
          "errorCode": -32884,
          "errorMessage": "Internal error"
        }
      statusCode: 500
    ...
```

##### template

This strategy gives ability to use incoming request data into mock response.
Implemented with package [text/template](https://pkg.go.dev/text/template).
Automatically preload incoming request into variable named `request`.

Parameters:

- `body` (mandatory) - sets the response body, must be valid `text/template` string;
- `statusCode` - HTTP-code of the response, the default value is `200`;
- `headers` - response headers.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: template
      body: >
        {
          "value-from-query": {{ .request.Query "value" }},
          "data-from-body": {{ default 10 .request.Json.data }}
        }
      statusCode: 200
    ...
```

##### uriVary

Uses different response strategies, depending on a path of a requested resource.

When receiving a request for a resource that is not defined in the parameters, the test will be considered failed.

Parameters:

- `uris` (mandatory) - a list of resources, each resource can be configured as a separate mock-service using any available request constraints and response strategies (see example)
- `basePath` - common base route for all resources, empty by default

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: uriVary
      basePath: /v2
      uris:
        /shelf/books:
          strategy: file
          filename: responses/books_list.json
          statusCode: 200
        /shelf/books/1:
          strategy: constant
          body: >
            {
              "error": "book not found"
            }
          statusCode: 404
    ...
```

##### methodVary

Uses various response strategies, depending on the request method.

When receiving a request with a method not defined in methodVary, the test will be considered failed.

Parameters:

- `methods` (mandatory) - a list of methods, each method can be configured as a separate mock-service using any available request constraints and response strategies (see example)

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: methodVary
      methods:
        GET:
          # nothing stops us from using `uriVary` strategy here
          # this way we can form different responses to different
          # method+resource combinations
          strategy: constant
          body: >
            {
              "error": "book not found"
            }
          statusCode: 404
        POST:
          strategy: nop
          statusCode: 204
    ...
```

##### sequence

With this strategy for each consequent request you will get a reply defined by a consequent nested strategy.

If no nested strategy specified for a request, i.e. arrived more requests than nested strategies specified, the test will be considered failed.

Parameters:

- `sequence` (mandatory) - list of nested strategies.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: sequence
      sequence:
        # Responds with a different text on each consequent request:
        # "1" for first call, "2" for second call and so on.
        # For 5th and later calls response will be 404 Not Found.
        - strategy: constant
          body: '1'
        - strategy: constant
          body: '2'
        - strategy: constant
          body: '3'
        - strategy: constant
          body: '4'
    ...
```

##### basedOnRequest

Allows multiple requests with same request path. Concurrent safe.

When receiving a request for a resource that is not defined in the parameters, the test will be considered failed.

Parameters:

- `uris` (mandatory) - a list of resources, each resource can be configured as a separate mock-service using any available request constraints and response strategies (see example)

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: basedOnRequest
      uris:
        - strategy: constant
          body: >
            {
              "ok": true
            }
          requestConstraints:
            - kind: queryMatches
              expectedQuery: "key=value1"
            - kind: pathMatches
              path: /request
        - strategy: constant
          body: >
            {
             "ok": true
            }
          requestConstraints:
            - kind: queryMatches
              expectedQuery: "key=value2"
            - kind: pathMatches
              path: /request
    ...
```

##### dropRequest

The strategy that by default drops the connection on any request. Used to emulate the network problems.

No parameters.

Example:

```yaml
  ...
  mocks:
    service1:
      strategy: dropRequest
    ...
```

#### Calls count

You can define, how many times each mock or mock resource must be called (using `uriVary`). If the actual number of calls is different from expected, the test will be considered failed.

Example:

```yaml
  ...
  mocks:
    service1:
      # must be called exactly one time
      calls: 1
      strategy: file
      filename: responses/books_list.json
  ...
```

```yaml
  ...
  mocks:
    service1:
      strategy: uriVary
      uris:
        /shelf/books:
          # must be called exactly one time
          calls: 1
          strategy: file
          filename: responses/books_list.json
  ...
```

## Shell scripts usage

When the test is ran, operations are performed in the following order:

1. Fixtures load
2. Mocks setup
3. beforeScript execute
4. HTTP-request sent
5. afterRequestScript execute
6. The checks are ran

### Script definition

To define the script you need to provide 2 parameters:

- `path` (mandatory) - string with a path to the script file.
- `timeout` - time in seconds, is responsible for stopping the script on timeout. The default value is `3`.

Example:

```yaml
  ...
  afterRequestScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will be equal 10s
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will be equal 10s
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will equal 3s
  ...
```

### Running a script with parameterization

When tests use parameterized requests, it's possible to use different scripts for each test run.

Example:

```yaml
  ...
  beforeScript:
    path: |
      ./cli_scripts/{{.file_name}}
  ...
  cases:
    - requestArgs:
        customer_id: 1
        customer_email: "customer_1_recalculate@example.com"
      responseArgs:
        200:
          rrr: 1
          in_transit: 1
      beforeScriptArgs:
        file_name: "cmd_recalculate_customer_1.sh"
```

## A DB query

After HTTP request execution you can run an SQL query to DB to check the data changes.
The response can contain several records. Those records are compared to the expected list of records.

### Test Format

You can use legacy style for run sql queries, like this:

```yaml
- name: my test
  ...
  dbQuery: >
    SELECT ...
  dbResponse:
    - ...
    - ...
```

But, for now, already acceptable style is:

```yaml
- name: my test
  ...
  dbChecks:
    - dbQuery: >
        SELECT ...
      dbResponse:
        - ...
```

With second variant, you can run any amount of needed queries, after test case runned.

*NOTE*: All mentioned below techniques are still work with both variants of query format.

### Query definition

Query is a SELECT that returns any number of strings.

- `dbQuery` - a string that contains an SQL query.

Example:

```yaml
  ...
  dbQuery: >
    SELECT code, purchase_date, partner_id FROM mark_paid_schedule AS m WHERE m.code = 'GIFT100000-000002'
  ...
```

### Definition of DB request response

The response is a list of JSON objects that the DB request should return.

- `dbResponse` - a string that contains a list of JSON objects.

Example:

```yaml
  ...
  dbResponse:
    - '{"code":"GIFT100000-000002","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"GIFT100000-000003","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"$matchRegexp(GIFT([0-9]{6})-([0-9]{6}))","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
```

As you can see in this example, you can use Regexp for checking db response body.

```yaml
  ...
  dbResponse:
    # empty list
```

### DB request parameterization

As well as with the HTTP request body, we can use parameterized requests.

Example:

```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
      - '{"code":"{{ .cert2 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
        cert2: "GIFT100000-000003"
```

When different tests contain different number of records, you can redefine the response for a specific test as a whole, while continuing to use a template with parameters in others.

Example:

```yaml
  ...
    dbQuery: >
      SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

    dbResponse:
      - '{"code":"{{ .cert1 }}","partner_id":1}'
  ...
    cases:
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"
      ...
      dbQueryArgs:
        fromDate: "2330-02-01"
        toDate: "2330-02-05"
      dbResponseFull:
        - '{"code":"GIFT100000-000002","partner_id":1}'
        - '{"code":"GIFT100000-000003","partner_id":1}'
```

### Ignoring ordering in DB response

You can use `ignoreDbOrdering` flag in `comparisonParams` section to toggle DB response ordering ignore feature.
This can be used to bypass using `ORDER BY` operators in query.

- `ignoreDbOrdering` - true/false value.

Example:

```yaml
  comparisonParams:
    ignoreDbOrdering: true
  ...
  dbQuery: >
    SELECT id, name, surname FROM users LIMIT 2
    
  dbResponse:
    - '{ "id": 2, "name": "John", "surname": "Doe" }'
    - '{ "id": 1, "name": "Jane", "surname": "Doe" }'
```

## JSON-schema

Use [file with schema](https://raw.githubusercontent.com/lansfy/gonkex/master/gonkex.json) to add syntax highlight to your favourite IDE and write Gonkex tests more easily.

It adds in-line documentation and auto-completion to any IDE that supports it.

Example in Jetbrains IDE:
![Example Jetbrains](https://i.imgur.com/oYuPuR3.gif)

Example in VSCode IDE:
![Example Jetbrains](https://i.imgur.com/hBIGjP9.gif)

### Setup in Jetbrains IDE

Download [file with schema](https://raw.githubusercontent.com/lansfy/gonkex/master/gonkex.json).
Open preferences File->Preferences
In Languages & Frameworks > Schemas and DTDs > JSON Schema Mappings

![Jetbrains IDE Settings](https://i.imgur.com/xkO22by.png)

Add new schema

![Add schema](https://i.imgur.com/XHw14GJ.png)

Specify schema name, schema file, and select Schema version: Draft 7

![Name, file, version](https://i.imgur.com/LfJfis0.png)

After that add mapping. You can choose from single file, directory, or file mask.

![Mapping](https://i.imgur.com/iFjm0Ld.png)

Choose what suits you best.

![Mapping pattern](https://i.imgur.com/WIK6sZW.png)

Save your preferences. If you done everything right, you should not see No JSON Schema in bottom right corner

![No Schema](https://i.imgur.com/zLqv1Zv.png)

Instead, you should see your schema name

![Schema Name](https://i.imgur.com/DDXdCO7.png)

### Setup is VSCode IDE

At first, you need to download YAML Language plugin
Open Extensions by going to Code(File)->Preferences->Extensions

![VSCode Preferences](https://i.imgur.com/X7bk5Kh.png)

Look for YAML and install YAML Language Support by Red Hat

![Yaml Extension](https://i.imgur.com/57onioF.png)

Open Settings by going to Code(File)->Preferences->Settings

Open Schema Settings by typing YAML:Schemas and click on *Edit in settings.json*
![Yaml link](https://i.imgur.com/IEwxWyG.png)

Add file match to apply the JSON on YAML files.

```
"yaml.schemas": {
  "C:\\Users\\Leo\\gonkex.json": ["*.gonkex.yaml"]          
}
```

In the example above the JSON schema stored in C:\Users\Leo\gonkex.json will be applied on all the files that ends with .gonkex.yaml
