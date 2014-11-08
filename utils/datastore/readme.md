# Grid Datastore
This is a simple in-memory datastore with expiring.

# Example

```
type Data struct {
    Answer int
}

ds := datastore.New(time.Millisecond * 500)
f := &Data{
    Answer: 42,
}

ds.Set("foo", f)

// get
o, err := ds.Get("foo")

// wait for ttl
o, err := ds.Get("foo")

// will return ErrKeyDoesNotExist as it has expired

```
