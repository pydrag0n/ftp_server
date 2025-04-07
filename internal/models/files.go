package models

type File struct {
    Filename string
    Path     string // Относительный путь от корня
    Size     int64
    Date     string
    IsDir    bool
    DefTheme bool
}


func (f *File) SetFilename(name string) {
    f.Filename = name
}

func (f *File) SetPath(path string) {
	f.Path = path
}

func (f *File) SetSize(size int64) {
	f.Size = size
}

func (f *File) SetDate(date string) {
	f.Date = date
}

func (f *File) SetIsDir(isDir bool) {
	f.IsDir = isDir
}

func (f *File) SetDefTheme(defTheme bool) {
	f.DefTheme = defTheme
}
