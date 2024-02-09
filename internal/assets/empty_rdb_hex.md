Here's the contents of an empty RDB file in base64: 

```
UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog==
```

Here's the same contents in hex: 

```
524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2
```

Here's a more readable [hexdump](https://opensource.com/article/19/8/dig-binary-files-hexdump) representation of the same file:

```
52 45 44 49 53 30 30 31  31 fa 09 72 65 64 69 73  |REDIS0011..redis|
2d 76 65 72 05 37 2e 32  2e 30 fa 0a 72 65 64 69  |-ver.7.2.0..redi|
73 2d 62 69 74 73 c0 40  fa 05 63 74 69 6d 65 c2  |s-bits.@..ctime.|
6d 08 bc 65 fa 08 75 73  65 64 2d 6d 65 6d c2 b0  |m..e..used-mem..|
c4 10 00 fa 08 61 6f 66  2d 62 61 73 65 c0 00 ff  |.....aof-base...|
f0 6e 3b fe c0 ff 5a a2                           |.n;...Z.|
```
