package settings

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
)

type stringLoader struct {
	src      string
	filename string
}

func NewStringLoader(src string) Loader {
	return &stringLoader{src, ""}
}

func (f *stringLoader) Parse() (map[string]string, error) {
	return f.parseData()
}

func (f *stringLoader) parseData() (res map[string]string, err error) {
	res = make(map[string]string)

	var dir string
	var file = f.filename
	if f.src == "" {
		if abs, e := filepath.Abs(f.filename); e == nil {
			dir, file = filepath.Split(abs)
		}
	}
	var data []byte
	if f.src == "" {
		data, err = ioutil.ReadFile(f.filename)
		if err != nil {
			return
		}
	} else {
		data = []byte(f.src)
	}

	p := &parser{
		filename: file,
	}
	p.r = bytes.NewReader(data)

	p.next()
	if p.ch == bom {
		p.next()
	}

	var key string
	for {
		tok, lit := p.scan()
		if tok == eof {
			break
		}
		if p.err != nil {
			return nil, p.err
		}
		switch tok {
		case comment:
			// pass
		case ident:
			if lit == "include" {
				// scan next token and expect it to be a string (filename)
				t, l := p.scan()
				if t != str {
					p.error("expecting filename")
					return nil, p.err
				}
				// resolve file path
				fname := l
				if !filepath.IsAbs(fname) {
					fname = filepath.Clean(filepath.Join(dir, l))
				}
				// parse the included file and append values to current result
				strLoader := &stringLoader{filename: fname}
				d, e := strLoader.parseData()
				if e != nil {
					return nil, e
				}
				for k, v := range d {
					res[k] = v
				}
				continue
			}
			if key == "" {
				key = lit
			} else {
				res[key] = lit
				key = ""
			}
		case equal:
			if key == "" {
				p.error("missing key")
				return nil, p.err
			}
		case number, float, str, duration, variable:
			if key == "" {
				p.error("missing key")
				return nil, p.err
			}
			res[key] = lit
			key = ""
		}
	}
	return
}
