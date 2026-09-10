package internal


const endpoint string = getPokeEndPoint("location-area?limit=3")

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}  {
		key: "0"
		val: []byte("This is the value for key 0")
	},{
		key: "20"
		val: []byte("Here is 20")
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value")
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime * 2
	cache := NewCache(baseTime)
	cash.Add("key", "val")

	_, ok := cache.Get("key")
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)
	
	_, ok := cache.Get("key")
	if !ok {
		t.Errorf("expected to find key")
		return
	}
}
