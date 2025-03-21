### Redis

Supports loading test data with fixtures for redis key/value storage.

List of supported data structures:

- Plain key/value
- Set
- Hash
- List
- ZSet (sorted set)

Fixture file example:

```yaml
inherits:
  - template1
  - template2
  - other_fixture
templates:
  keys:
    - $name: parentKeyTemplate
      values:
        baseKey:
          expiration: 1s
          value: 1
    - $name: childKeyTemplate
      $extend: parentKeyTemplate
      values:
        otherKey:
          value: 2
  sets:
    - $name: parentSetTemplate
      expiration: 10s
      values:
        - value: a
    - $name: childSetTemplate
      $extend: parentSetTemplate
      values:
        - value: b
  hashes:
    - $name: parentHashTemplate
      values:
        - key: a
          value: 1
        - key: b
          value: 2
    - $name: childHashTemplate
      $extend: parentHashTemplate
      values:
        - key: c
          value: 3
        - key: d
          value: 4
  lists:
    - $name: parentListTemplate
      values:
        - value: 1
        - value: 2
    - $name: childListTemplate
      values:
        - value: 3
        - value: 4
  zsets:
    - $name: parentZSetTemplate
      values:
        - value: 1
          score: 2.1
        - value: 2
          score: 4.3
    - $name: childZSetTemplate
      value:
        - value: 3
          score: 6.5
        - value: 4
          score: 8.7
databases:
  1:
    keys:
      $extend: childKeyTemplate
      values:
        key1:
          value: value1
        key2:
          expiration: 10s
          value: value2
    sets:
      values:
        set1:
          $extend: childSetTemplate
          expiration: 10s
          values:
            - value: a
            - value: b
        set3:
          expiration: 5s
          values:
            - value: x
            - value: y
    hashes:
      values:
        map1:
          $extend: childHashTemplate
          values:
            - key: a
              value: 1
            - key: b
              value: 2
        map2:
          values:
            - key: c
              value: 3
            - key: d
              value: 4
    lists:
      values:
        list1:
          $extend: childListTemplate
          values:
            - value: 1
            - value: 100
            - value: 200
    zsets:
      values:
        zset1:
          $extend: childZSetTemplate
          values:
            - value: 5
              score: 10.1
  2:
    keys:
      values:
        key3:
          value: value3
        key4:
          expiration: 5s
          value: value4
```
