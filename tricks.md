- Use bytes.FieldsFunc() to abstract splitting of bytes. Its more general than bytes.Split() *only if* it is utf-8

- string(os.PathSeparator) should replace `/`