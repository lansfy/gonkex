### Aerospike

Fixtures for Aerospike are also supported. Add `DbType: fixtures.Aerospike` to runner's configuration if gonkey is used as library.

Fixtures files format is a bit different, yet the same basic principles applies:

```yaml
sets:
  set1:
    key1:
      bin1: "value1"
      bin2: 1
    key2:
      bin1: "value2"
      bin2: 2
      bin3: 2.569947773654566473
  set2:
    key1:
      bin4: false
      bin5: null
      bin1: '"'
    key2:
      bin1: "'"
      bin5:
        - 1
        - '2'
```

Fixtures templates are also supported:

```yaml
templates:
  base_tmpl:
    bin1: value1
  extended_tmpl:
    $extend: base_tmpl
    bin2: value2

sets:
  set1:
    key1:
      $extend: base_tmpl
      bin1: overwritten
  set2:
    key1:
      $extend: extended_tmpl
      bin2: overwritten
```

Expressions are currently not supported.
