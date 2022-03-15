# util

To publish new module:
```
git commit .
git tag --list
git tag v0.1.xx
git push
git push origin v0.1.xx
```

To use new module:
```
cd ../complaints
go get github.com/skypies/util@v0.1.xx
```
