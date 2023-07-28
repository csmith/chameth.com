---
title: Extracting a Zip file
group: Go
---

You need to know the length of the zip file up-front, so it's generally easiest to read it into a `[]byte` and then:

```go
zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))  
if err != nil {  
   panic(err)  
}  
  
for i := range zr.File {  
   func(f *zip.File) {  
      if f.FileInfo().IsDir() { 
         return  
      }  
  
      file, err := f.Open()  
      if err != nil {  
         panic(err)  
      }  
      defer file.Close()  
  
      bs, err := io.ReadAll(file)  
      if err != nil {  
         panic(err)  
      }  
  
      target := filepath.Join(dir, f.Name)
  
      if err := os.MkdirAll(filepath.Dir(target), os.FileMode(0755)); err != nil {  
         panic(err)  
      }  
  
      if err := os.WriteFile(target, bs, os.FileMode(0644)); err != nil {  
         panic(err)  
      }  
   }(zr.File[i])  
}
```

This assumes the zip is trusted --- user supplied zip files could have malicious paths.
Proper error handling is left as an exercise for the reader.