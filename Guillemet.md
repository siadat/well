# Guillemets

## Why not single (') and double (") quotes?
The problem with single quote and double quote is that nesting them becomes unreadable and prone to mistakes when writing.
It is not possible to tell if the following is correct with just looking:

```
sh -c 'sh -c "echo \'a string with \\\\\'single-quotes\\\\\' and a string with \\"double-quotes\\"\'"'
```

But this one is much easier to read:
```
sh -c ‹sh -c «echo ‹${a} and ${b}›»›
```

You can get that using:
```
a="a string with 'single-quotes'" \
b='a string with "double-quotes"' \
  guillemets render --input 'sh -c ‹sh -c «echo ‹${a} and ${b}›»›'
```

Note that when open and close characters are different, we are able to nest them without having to escape them.

## Why not “this” and ‘this’ (MS Word style)?

- They can be confused with existing quotes, because they look very similar to them.
- The open and close ones are not distinguishable in small font.

## Guillemets are not on my keyboard?

Some keyboard layouts include guillemets, some don't.
It is not a big issue, because these characters can be copy-and-pasted it when you need it.
Writing programs is more like painting that performing live music.
