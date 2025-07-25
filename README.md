# Gonkex: testing automation tool

[![Go Reference](https://pkg.go.dev/badge/github.com/lansfy/gonkex.svg)](https://pkg.go.dev/github.com/lansfy/gonkex)
[![coverage](https://raw.githubusercontent.com/lansfy/gonkex/refs/heads/badges/.badges/master/coverage.svg)](https://github.com/lansfy/gonkex/blob/master/.testcoverage.yml)

*NOTE*: You can find a utility that allows you to run gonkex scripts [here](https://github.com/lansfy/gonkex-cli) (in the releases section).

Gonkex will test your services using their API. It can bomb the service with prepared requests and check the responses. Test scenarios are described in YAML-files.

Capabilities:

- works with REST/(JSON,XML,YAML) API
- provides [declarative mocks](#mocks) for external services
- seeds the database with [fixtures data](#fixtures) (supports PostgreSQL, MySQL, Sqlite, TimescaleDB, MariaDB, SQLServer, ClickHouse, Aerospike, MongoDB, Redis)
- [execute and verify database queries](#a-db-query) to check test outcomes
- run as a [standalone tool](https://github.com/lansfy/gonkex-cli/) or as a [library](#using-gonkex-as-a-library) alongside your unit tests
- stores the results as an [Allure](https://allurereport.org/) report
- there is a [JSON-schema](#json-schema) to add autocomplete and validation for Gonkex YAML files

## Table of contents

- [Table of contents](#table-of-contents)
- [Using Gonkex as a library](#using-gonkex-as-a-library)
- [Test scenario example](#test-scenario-example)
- [HTTP-request](#http-request)
- [HTTP-response](#http-response)
- [Test status](#test-status)
- [Retry policy](#retry-policy)
- [Customizing a comparison](#customizing-a-comparison)
- [Pattern matching](#pattern-matching)
   * [$matchRegexp](#matchregexp)
   * [$matchTime](#matchtime)
      + [Basic Format Matching](#basic-format-matching)
      + [`accuracy` parameter](#accuracy-parameter)
      + [`value` parameter](#value-parameter)
      + [`timezone` parameter](#timezone-parameter)
   * [$matchArray](#matcharray)
      + [$matchArray(pattern)](#matcharraypattern)
      + [$matchArray(subset+pattern)](#matcharraysubsetpattern)
      + [$matchArray(pattern+subset)](#matcharraypatternsubset)
- [Delays](#delays)
- [Variables](#variables)
   * [Assignment](#assignment)
      + [In the description of the test](#in-the-description-of-the-test)
      + [From the response of the previous test](#from-the-response-of-the-previous-test)
      + [From the response body of currently running test](#from-the-response-body-of-currently-running-test)
      + [From environment variables or from env-file](#from-environment-variables-or-from-env-file)
      + [From cases](#from-cases)
- [multipart/form-data requests](#multipartform-data-requests)
   * [Form](#form)
   * [File upload](#file-upload)
- [Fixtures](#fixtures)
   * [Record templates](#record-templates)
   * [Record inheritance](#record-inheritance)
   * [Expressions](#expressions)
   * [Deleting data from tables](#deleting-data-from-tables)
- [Mocks](#mocks)
   * [Running mocks while using Gonkex as a library](#running-mocks-while-using-gonkex-as-a-library)
   * [Mocks definition in the test file](#mocks-definition-in-the-test-file)
      + [Request constraints (requestConstraints)](#request-constraints-requestconstraints)
         - [nop](#nop)
         - [methodIs](#methodis)
         - [headerIs](#headeris)
         - [pathMatches](#pathmatches)
         - [queryMatches](#querymatches)
         - [queryMatchesRegexp](#querymatchesregexp)
         - [bodyMatchesText](#bodymatchestext)
         - [bodyMatchesJSON](#bodymatchesjson)
         - [bodyMatchesXML](#bodymatchesxml)
         - [bodyMatchesYAML](#bodymatchesyaml)
         - [bodyJSONFieldMatchesJSON](#bodyjsonfieldmatchesjson)
      + [Response strategies (strategy)](#response-strategies-strategy)
         - [nop](#nop-1)
         - [constant](#constant)
         - [file](#file)
         - [template](#template)
         - [uriVary](#urivary)
         - [methodVary](#methodvary)
         - [sequence](#sequence)
         - [basedOnRequest](#basedonrequest)
         - [dropRequest](#droprequest)
      + [Calls count](#calls-count)
- [Shell scripts usage](#shell-scripts-usage)
   * [Script definition](#script-definition)
   * [Running a script with parameterization](#running-a-script-with-parameterization)
- [A DB query](#a-db-query)
   * [Query definition](#query-definition)
   * [Definition of DB request response](#definition-of-db-request-response)
   * [DB request parameterization](#db-request-parameterization)
   * [Ignoring ordering in DB response](#ignoring-ordering-in-db-response)
- [JSON-schema](#json-schema)

## Using Gonkex as a library

To integrate functional and native Go tests and run them together, use Gonkex as a library.

Create a test file, for example `func_test.go`.

Import Gonkex as a dependency to your project in this file and create a test function:

```go
package test

import (
	"testing"

	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/storage"
	"github.com/lansfy/gonkex/storage/addons/sqldb"
	"github.com/lansfy/gonkex/runner"
)

func init() {
	// Optional helper function which registers "gonkex-filter" flag that allows users
	// to filter which test files are executed during a test run.
	// For example, go test -gonkex-filter="mytest.yaml"
	runner.RegisterFlags()
}

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

Externally written storage may be used for loading test data, if Gonkex used as a library.
To start using the custom storage, you need to import the custom module, that contains implementation of storage.StorageInterface interface.
For example, the following NoSQL databases are currently supported as custom modules:
- Aerospike ([storage/addons/aerospike](https://github.com/lansfy/gonkex/tree/master/storage/addons/aerospike))
- MongoDB ([storage/addons/mongo](https://github.com/lansfy/gonkex/tree/master/storage/addons/mongo))
- Redis ([storage/addons/redis](https://github.com/lansfy/gonkex/tree/master/storage/addons/redis))

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

`method` - a parameter for HTTP request type (e.g. `GET`, `POST`, `DELETE` and so on)

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

*NOTE*: An attempt is considered successful if the actual response matches the expected response.

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

The following fields are supported:

`attempts` - an integer indicating the number of times that Gonkex will retry the test request in the event assertions fail.

`delay` - string containing the waiting time after unsuccessful completion of the test.

`successInRow` - parameter defines the required number of successful test passes for the test to be recognized as successful. And all these successful runs must be consecutive. Default value is 1.

## Customizing a comparison

After receiving a response from the service, the test compares the body of the received response with the body specified in the test.
By default, only the values of the fields listed in the test body are compared, but you can control the comparison procedure by using boolean flags in the `comparisonParams` section.
The following flags are supported:

- `ignoreValues` - if `true`, ignores differences in values and only checks the structure

- `ignoreArraysOrdering` - if `true`, considers arrays equal regardless of the order of elements

- `disallowExtraFields` - if `true`, fails the comparison if extra fields exist in the compared structure

All flags are set to `false` by default.

Example:
```yaml
 - name: compare flag example
   ...
   comparisonParams:
     ignoreValues: true
     ignoreArraysOrdering: true
     disallowExtraFields: true
```

## Pattern matching

The pattern matching is a feature in Gonkex that allows you to validate response, mock request, DB query results using some pattern (like regular expressions) instead of exact matching.
This is especially useful when you testing dynamic or unpredictable parts of data (like timestamps, UUIDs, or random tokens).

### $matchRegexp

The basic syntax for using `$matchRegexp` is:

```yaml
$matchRegexp(regular_expression)
```

where `regular_expression` is a valid Go [regular expression](https://pkg.go.dev/regexp/syntax) pattern.

Example:

```yaml
- name: WHEN order information is requested, service MUST return valid order data
  method: GET
  path: /api/orders/12345
  response:
    200: >
      {
        "order_id": "$matchRegexp(^\\d{5,7}$)",
        "created_at": "$matchRegexp(^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z$)",
        "status": "$matchRegexp(pending|processing|shipped|delivered)",
        "total_amount": 1299.99,
        "transaction_id": "$matchRegexp(^txn_[a-zA-Z0-9]{24}$)",
        "tracking_number": "$matchRegexp(^(TR\\d{10})?$)"
      }
```

*NOTE*: If you want to match the entire string, use `^` at the beginning and `$` at the end of your pattern. 

### $matchTime

The `$matchTime` function is allows you to validate timestamp strings in response, mock request, DB query results according to specific time format patterns.
Unlike the more general `$matchRegexp`, `$matchTime` is designed specifically for time validation.
This feature is used when you cannot specify the exact time (for example, the time in the response depends on the current time).

The basic syntax for using `$matchTime` is:

```yaml
$matchTime(format_string[, parameter=value][, ...])
```

where:
- `format_string` is a valid [Go time format](https://pkg.go.dev/time#pkg-constants) or [strftime time format](https://pkg.go.dev/github.com/ncruces/go-strftime#pkg-overview) pattern
- optional parameters can be added to customize the time matching behavior

#### Basic Format Matching

The simplest usage of `$matchTime` validates that a timestamp string matches the specified format:

```yaml
  ...
  response:
    200: >
      {
        "id": "12345",
        "created_at": "$matchTime(2006-01-02T15:04:05Z07:00)",
        "updated_at": "$matchTime(%Y-%m-%dT%H:%M:%S%z)",
        "event_date": "$matchTime(Jan 2, 2006)",
        "scheduled_time": "$matchTime(%H:%M:%S)"
      }
  ...
```

*NOTE*: For consistency, try to stick to one format style (Go or Strftime format) in all tests.

#### `accuracy` parameter

Defines the acceptable time difference when using the `value` parameter:

- `accuracy=duration` - sets a bidirectional time window (e.g., `accuracy=5m` for �5 minutes)
- `accuracy=+duration` - sets a forward-only time window (e.g., `accuracy=+10m` for 0 to +10 minutes)
- `accuracy=-duration` - sets a backward-only time window (e.g., `accuracy=-10m` for -10 to 0 minutes)

By default, `accuracy` is set to �5 minutes when using any `value`.

```yaml
response:
  200: >
    {
      "timestamp_precise": "$matchTime(%Y-%m-%d %H:%M:%S, value=now, accuracy=1m)",
      "timestamp_future": "$matchTime(%Y-%m-%d %H:%M:%S, value=now, accuracy=+30m)",
      "timestamp_past": "$matchTime(%Y-%m-%d %H:%M:%S, value=now, accuracy=-30m)"
    }
```

*NOTE*: `duration` should be defined using Go [time duration string](https://pkg.go.dev/time#ParseDuration). For convenience, days (`d`) and weeks (`w`) are also supported.

#### `value` parameter

Allows you to specify an expected time value to match against:

- `value=now` or `value=now()` - matches times around the current system time
- `value=now�offset` - matches times offset from the current time (e.g., `value=now-1h`, `value=now+30m`)
- `value=specific_time`- matches a specific time in the same format as the pattern (e.g., `value=25-12-2023 10:20:30` for format `%d-%m-%Y %H:%M:%S`)

```yaml
response:
  200: >
    {
      "last_login": "$matchTime(%Y-%m-%d %H:%M:%S, value=now-1h)",
      "next_scheduled": "$matchTime(%Y-%m-%d %H:%M:%S, value=now+24h)",
      "specific_date": "$matchTime(%d-%m-%Y %H:%M:%S, value=25-12-2023 10:20:30)"
    }
```

*NOTE*: `offset` should be defined using Go [time duration string](https://pkg.go.dev/time#ParseDuration). For convenience, days (`d`) and weeks (`w`) are also supported.

#### `timezone` parameter

Allows you to specify timezone for values without specified timezone:

- `timezone=local` - use local timezone (default)
- `timezone=utc` - use UTC timezone

### $matchArray

The `$matchArray` feature allows you to validate that all elements in an array match a specific pattern. This is especially useful when:

- you don't know exactly how many elements will be in the array;
- all elements in the array should follow the same pattern or structure;
- you want to avoid repetitive pattern definitions for large arrays.

#### $matchArray(pattern)

To use `$matchArray`, you need to define an array with exactly two elements:

- the literal string `$matchArray(pattern)`;
- a pattern object that defines what each array element should match.

Example:

```yaml
- name: WHEN orders information is requested, service MUST return valid orders data
  method: GET
  path: /api/orders

  response:
    200: >
      {
        "user": "testuser",
        "orders": [
          "$matchArray(pattern)",
          {
            "order_id": "$matchRegexp(^ORDER[0-9]{4}$)",
            "amount": "$matchRegexp(^[0-9]+\\.?[0-9]*$)",
            "status": "$matchRegexp(pending|processing|completed)"
          }
        ]
      }
```

This pattern will match arrays of any length, as long as all elements follow the specified structure.

#### $matchArray(subset+pattern)

In this mode:

- the first element in your test array must be the literal string `$matchArray(subset+pattern)`;
- the last element defines the pattern that any additional elements in the response array must match;
- all elements between these two (the subset) are treated as required initial elements that must appear at the beginning of the response array in the exact order specified;
- after matching these initial elements, any remaining elements in the response array must match the pattern defined in the last element.

*NOTE*: you still can use the `ignoreArraysOrdering` parameter with `$matchArray(subset+pattern)`.
When set to `true`, this parameter allows the subset elements to appear anywhere in the array, not just at the beginning, while still maintaining the pattern matching for additional elements.

#### $matchArray(pattern+subset)

In this mode:

- the first element in your test array must be the literal string `$matchArray(pattern+subset)`;
- the second element defines the pattern that any leading elements in the response array must match;
- all elements after these two (the subset) are treated as required final elements that must appear at the end of the response array in the exact order specified;
- the beginning of the response array must contain zero or more elements that match the pattern defined in the second element.

```yaml
- name: WHEN products are requested, service MUST return regular products followed by featured products
  method: GET
  path: /api/products
  response:
    200: >
      {
        "products": [
          "$matchArray(pattern+subset)",
          {
            "product_id": "$matchRegexp(^PROD-[A-Z0-9]{6}$)",
            "price": "$matchRegexp(^\\d+\\.\\d{2}$)",
            "featured": false
          },
          {
            "product_id": "FEATURED-001",
            "price": "29.99",
            "featured": true
          },
          {
            "product_id": "FEATURED-002",
            "price": "49.99",
            "featured": true
          }
        ]
      }
```

*NOTE*: you still can use the `ignoreArraysOrdering` parameter with `$matchArray(pattern+subset)`.
When set to `true`, this parameter allows the subset elements to appear anywhere in the array, not just at the end, while still maintaining the pattern matching for additional elements.

## Delays

`pause` - amount of time that the test should wait before executing.

`afterRequestPause` - amount of time that the test should wait after executing. It is important to note that this wait is part of the request test, i.e. all checks and mocks constraints will be checked after the wait is complete.

This delays should be defined using Go [time duration string](https://pkg.go.dev/time#ParseDuration).

## Variables

You can use variables in the description of the test, the following fields are supported:

- method
- description
- path
- query
- headers
- request
- response
- response headers
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
  responseHeaders:
    200:
      Some-Header: "{{ $respHeader }}"
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
    200:
      id: ""                      # store whole text body to variable

# if the response is JSON
- name: "get_last_post_info"
  ...
  variables_to_set:
    200:
      id: "id"
      title: "title"
      authorId: "author_info.id"  # get nested json field (any nesting levels are supported)
      wholeBody: ""               # empty path tells to put whole response body to variable
```

All paths must be specified in [gjson format](https://github.com/tidwall/gjson/blob/master/SYNTAX.md). You can use the [GJSON Playground](https://gjson.dev) to experiment with the syntax online.

It is also possible to retrieve values from the headers and cookies of response. To do this, specify the prefix `header:` or `cookie:` in the path, respectively. For example,

```yaml
- name: "get_data_from_last_response"
  ...
  variables_to_set:
    302:
      newLocation: "header:Location"    # get value from "Location" header and put to newLocation variable
      sessionId: "cookie:session_id"    # get value from "session_id" cookie and put to sessionId variable
      authorId: "body:author_info.id"   # optional "body:" prefix allows to get value from body
```

#### From the response body of currently running test

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

To seed the DB before the test, Gonkex uses fixture files.

File example:

```yaml
# fixtures/comments.yml
inherits:
  - another_fixture
  - yet_another_fixture

tables:
  posts:
    - id: 100
      title: New post
      text: Post text
      author: Jane Dow
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - id: 110
      title: Morning digest
      text: Text
      author: Apple Seed
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  comments:
    - post_id: 100
      content: A comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

    - post_id: 110
      content: Another comment...
      author_name: John Doe
      author_email: john@doe.com
      created_at: 2016-01-01 12:30:12
      updated_at: 2016-01-01 12:30:12

  another_table:
    ...
  ...
```

Records in fixtures can use templates and inherit.

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

*NOTE*: Record inheritance only works with different fixture files. It's not possible to declare inheritance within one file.

### Expressions

When you need to write an expression execution result to the DB and not a static value, you can use `$eval(...)` construct.
Everything inside the brackets will be inserted into the DB as raw, non-escaped data.
This way, within `$eval()` you can write everything you would in a regular query.

For instance, this construct allows the insertion of current date and time as a field value:

```yaml
tables:
  comments:
    - created_at: $eval(NOW())
```

### Deleting data from tables

To clear the table before the test put square brackets next to the table name.

Example:

```yaml
# fixtures/empty_posts_table.yml
tables:
  # cleanup posts table
  posts: []
```

## Mocks

In order to imitate responses from external services, use mocks.

A mock is a web server that is running on-the-fly, and is populated with certain logic before the execution of each test.
The logic defines what the server responses to a certain request. It's defined in the test file.

### Running mocks while using Gonkex as a library

Before running tests, all planned mocks are started.
It means that Gonkex spins up the given number of servers and each one of them gets a random port assigned.

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

##### methodIs

Checks that the request method corresponds to the expected one.

Parameters:

- `method` (mandatory) - string to compare the request method to.

For the most commonly used methods, there are also short variants that do not require the `method` parameter:

- `methodIsGET`
- `methodIsPOST`
- `methodIsPUT`
- `methodIsDELETE`

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: methodIs
          method: PUT
    ...
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

It is also possible to specify a regular expression using `$matchRegexp` in the `value` field.

Examples:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          value: application/json
    ...
    service2:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          regexp: ^(application/json|text/plain)$
    ...
    service3:
      requestConstraints:
        - kind: headerIs
          header: Content-Type
          value: "$matchRegexp(^(application/json|text/plain)$)"
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
    ...
    service2:
      requestConstraints:
        - kind: pathMatches
          regexp: ^/api/v1/test/.*$
    ...
```

##### queryMatches

Checks that the GET request parameters correspond to the ones defined in the `query` parameter.

Parameters:

- `query` (mandatory) - a list of parameters to compare the parameter string to. The order of parameters is not important.

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
          query:  key1=value1&key2=value2&key1=value11
    ...
```

*NOTE*: For backward compatibility, the use of the `expectedQuery` parameter instead of `query` is also supported.

##### queryMatchesRegexp

Expands `queryMatches` so it can be used with regexp pattern matching.

Parameters:

- `query` (mandatory) - a list of parameters to compare the parameter string to. The order of parameters is not important.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        # works similarly to queryMatches with an addition of $matchRegexp usage
        - kind: queryMatchesRegexp
          query:  key1=value1&key2=$matchRegexp(\\d+)&key1=value11
    ...
```

*NOTE*: For backward compatibility, the use of the `expectedQuery` parameter instead of `query` is also supported.

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
    ...
    service2:
      requestConstraints:
        - kind: bodyMatchesText
          regexp: (HeroNameAndFriends)
    ...
```

##### bodyMatchesJSON

Checks that the request body is JSON, and it corresponds to the JSON defined in the `body` parameter.

Parameters:

- `body` (mandatory) - expected JSON (all keys on all levels defined in this parameter must be present in the request body);
- `comparisonParams` - section allows you to customize the comparison process.

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

##### bodyMatchesXML

Checks that the request body is XML, and it matches to the XML defined in the `body` parameter.

Parameters:

- `body` (mandatory) - expected XML;
- `comparisonParams` - section allows you to customize the comparison process.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesXML
          body: |
            <Person>
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

##### bodyMatchesYAML

Checks that the request body is YAML, and it matches to the YAML defined in the `body` parameter.

Parameters:

- `body` (mandatory) - expected YAML;
- `comparisonParams` - section allows you to customize the comparison process.

Example:

```yaml
  ...
  mocks:
    service1:
      requestConstraints:
        - kind: bodyMatchesYAML
          body: |
              FullName: "Harry Potter"
              Email:
                work: "hpotter@hog.gb"
                home: "hpotter@gmail.com"
              Addr: "4 Privet Drive"
              Group:
                - Hexes
                - Jinxes
    ...
```

##### bodyJSONFieldMatchesJSON

When request body is JSON, checks that value of particular JSON-field is string-packed JSON
that matches to JSON defined in `value` parameter.

Parameters:

- `path` (mandatory) - path to string field, containing JSON to check;
- `value` (mandatory) - expected JSON;
- `comparisonParams` - section allows you to customize the comparison process.

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
      body: |
        {
          "value-from-query": "{{ .request.Query "some_value" }}",
          "data-from-body": "{{ .request.Json.data }}"
        }
      statusCode: 200
    ...
```

##### uriVary

Uses different response strategies, depending on a path of a requested resource.

When receiving a request for a resource that is not defined in the parameters, the test will be considered failed.

Parameters:

- `uris` (mandatory) - a list of resources, each resource can be configured as a separate mock-service using any available request constraints and response strategies (see example);
- `basePath` - common base route for all resources, empty by default.

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

When receiving a request with a method not defined in `methodVary`, the test will be considered failed.

Parameters:

- `methods` (mandatory) - a list of methods, each method can be configured as a separate mock-service using any available request constraints and response strategies (see example).

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
        # For 5th and later calls response will be "200 OK" with empty body and fail the test case.
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

Allows multiple requests with same request path.
When receiving a request to mock, all elements in the `uris` list are sequentially passed through and the first element is returned, all checks (`requestConstraints`) of which will pass successfully.
If no such element is found, the test will be considered failed. This stratagy is concurrent safe.

Parameters:

- `uris` (mandatory) - a list of resources, each resource can be configured as a separate mock-service using any available request constraints and response strategies (see example).

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
            - kind: pathMatches
              path: /request
            - kind: queryMatches
              query: "key=value1"
        - strategy: constant
          body: >
            {
             "ok": true
            }
          requestConstraints:
            - kind: pathMatches
              path: /request
            - kind: queryMatches
              query: "key=value2"
    ...
```

##### dropRequest

When any request is received, this strategy drops the connection to the client. Used to emulate the network problems.

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

You can define, how many times each mock or mock resource must be called. If the actual number of calls is different from expected, the test will be considered failed.

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
4. pause before request
5. HTTP-request sent
6. afterRequestPause
7. afterRequestScript execute
8. The checks are ran

### Script definition

To define the script you need to provide 2 parameters:

- `path` (mandatory) - string with a path to the script file.
- `timeout` - time is responsible for stopping the script on timeout. Should be specified in Go [time duration string](https://pkg.go.dev/time#ParseDuration) or in seconds. The default value is `3s`.

Example:

```yaml
  ...
  afterRequestScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will be equal 500 milliseconds (defined as duration string)
    timeout: 500ms
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # the timeout will be equal 10 seconds (if a number is specified, it is assumed to be the number of seconds)
    timeout: 10
  ...
```

```yaml
  ...
  beforeScript:
    path: './cli_scripts/cmd_recalculate.sh'
    # default timeout equal 3 seconds
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

Use the following syntax to query the database:

```yaml
- name: my test
  ...
  dbChecks:
    - dbQuery: "SELECT ..."   # first query
      dbResponse:
        - ...
        - ...
    - dbQuery: "SELECT ..."   # second query
      dbResponse:
        - ...
        - ...
      comparisonParams:       # you can add a comparisonParams section to customize the comparison
        ignoreArraysOrdering: true
        disallowExtraFields: true
    - ....
```

This syntax allows any number of queries to be executed after the test case is complete.

You can also use legacy style for run sql queries (but this method only allows you to execute one query), like this:

```yaml
- name: my test
  ...
  dbQuery: "SELECT ..."
  dbResponse:
    - ...
    - ...
```

*NOTE*: All mentioned below techniques are still work with both variants of query format.

### Query definition

Query is a SELECT that returns any number of records.

- `dbQuery` - a string that contains an SQL query.

Example:

```yaml
  ...
  dbQuery: "SELECT code, purchase_date, partner_id FROM mark_paid_schedule AS m WHERE m.code = 'GIFT100000-000002'"
  ...
```

### Definition of DB request response

The response is a list of records in JSON format that the DB query should return.

- `dbResponse` - list of strings containing JSON objects.

Example:

```yaml
  ...
  dbResponse:
    - '{"code":"GIFT100000-000002","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"GIFT100000-000003","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
    - '{"code":"$matchRegexp(^GIFT([0-9]{6})-([0-9]{6})$)","purchase_date":"2330-02-02T13:15:11.912874","partner_id":1}'
```

As you can see in this example, you can use Regexp for checking DB response body.

To show that the query returns no records, you can specify an empty list in `dbResponse`. For example,

```yaml
  ...
  dbResponse: []   # empty list
```

Gonkex allows you to add a `comparisonParams` section to the database query parameters to customize the result comparison process.

### DB request parameterization

As well as with the HTTP request body, we can use parameterized requests.

Example:

```yaml
  ...
  dbChecks:
    - dbQuery: >
        SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

      dbResponse:
        - '{"code":"{{ .cert1 }}","partner_id":1}'
        - '{"code":"{{ .cert2 }}","partner_id":1}'
  ...
  cases:
    - dbQueryArgs:
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
  dbChecks:
    - dbQuery: >
        SELECT code, partner_id FROM mark_paid_schedule AS m WHERE DATE(m.purchase_date) BETWEEN '{{ .fromDate }}' AND '{{ .toDate }}'

      dbResponse:
        - '{"code":"{{ .cert1 }}","partner_id":1}'
  ...
  cases:
    - dbQueryArgs:
        fromDate: "2030-02-01"
        toDate: "2030-02-05"
      dbResponseArgs:
        cert1: "GIFT100000-000002"

    - dbQueryArgs:
        fromDate: "2030-02-01"
        toDate: "2030-02-05"
      dbResponseFull:
        - '{"code":"GIFT100000-000002","partner_id":1}'
        - '{"code":"GIFT100000-000003","partner_id":1}'
```

### Ignoring ordering in DB response

Gonkex allows you to add a `comparisonParams` section to the database query parameters to customize the result comparison process.
For example, you can specify the `ignoreArraysOrdering` flag to ignore the order of records when comparing.
This can be used to bypass the use of `ORDER BY` operators in a query.

Example:

```yaml
  ...
  dbChecks:
    - dbQuery: "SELECT id, name, surname FROM users LIMIT 2"
      dbResponse:
        - '{ "id": 2, "name": "John", "surname": "Doe" }'
        - '{ "id": 1, "name": "Jane", "surname": "Doe" }'

      comparisonParams:
        ignoreArraysOrdering: true
```

## JSON-schema

Use [file with schema](https://raw.githubusercontent.com/lansfy/gonkex/master/schema/gonkex.json) to add syntax highlight to your favourite IDE and write Gonkex tests more easily.
It adds in-line documentation and auto-completion to any IDE that supports it.
The [following article](https://github.com/lansfy/gonkex/tree/master/schema) describes how to add schema to your IDE.
