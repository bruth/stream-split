# Stream Split

[`split`](http://man7.org/linux/man-pages/man1/split.1.html) but streams each chunk to the provided command.

```
stream-split -lines 500 data.txt -- wc -l
```

This will split up `data.txt` into 500 line chunks and execute `wc -l` for each chunk.
