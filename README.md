# steganography

A Go application to encode and decode images with strings.

## Usage - Encode
```
go run steganography.go -s 'This is a test!' -i './images/UTImage.bmp'    
word: This is a test!
decode: false
(0,0)-(800,464)
800
464
```

## Usage - Decode
```
go run steganography.go -i './images/outimage.bmp' -d
word: foo
decode: true
(0,0)-(800,464)
800
464
This is a test!
```