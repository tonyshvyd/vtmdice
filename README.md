# VtM 5e dice thrower

CLI app throwing dices for VtM 5e.

# How to run

Build and run using golang

```
go build
./vtmdice
```

# Input commands

- `q` - exit
- `5 2` - throw 5 dices, 2 of them are hunger dices
- `5 2 3` - same as above but against difficulty 3. Counts failures or success.
- `rc` - Rouse Check. Shortcut for `1 1`
- `bs` - Blood Surge. Makes rc and rolls 2 additional dices
- `r - 1 2 3` - reroll given dices.
- `r 2` - reroll number of any dices.

# Credits

VtM is owned by https://www.worldofdarkness.com/dark-pack
