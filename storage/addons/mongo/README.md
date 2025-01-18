### MongoDB storage

To connect to MongoDB, you need to:
- For the Package-version: when configuring the runner, set DbType: fixtures.Mongo and pass the MongoDB client as Mongo: {mongo client}.

The format of fixture files for MongoDB:
```yaml
collections:
  collection1:
    - field1: "value1"
      field2: 1
    - field1: "value2"
      field2: 2
      field3: 2.569947773654566
  collection2:
    - field4: false
      field5: null
      field1: '"'
    - field1: "'"
      field5:
        - 1
        - '2'
```

If you are using different databases:
```yaml
collections:
  database1.collection1:
    - f1: value1
      f2: value2

  database2.collection2:
    - f1: value3
      f2: value4

  collection3:
    - f1: value5
      f2: value6
```

The `eval` operator is not supported.
