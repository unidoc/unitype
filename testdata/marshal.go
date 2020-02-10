/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

// TODO(gunnsth): which is a better name byteReader or streamReader ?

// unmarshaller is the interface that tables implement for loading the full table data
// from the reader.
type unmarshaller interface {
	// Unmarshal fully unmarshals the full table data.
	Unmarshal(r *byteReader) error
}

// marshaller is the interface that tables implement for writing out the tabular information
// from the writer. If the table is only preloaded then writes the unchanged information out,
// otherwise fully unmarhals the possibly changed data.
type marshaller interface {
	// Marshal writes the table data out, cached if preloaded, otherwise fully unmarshaling
	// the underlying data back out.
	Marshal(w *byteWriter) error
}

// interface declarations.
// TODO(gunnsth): All the major tables should implement preloader, unmarshaller and marshaller.
/*
var (
	_ marshaller = &offsetTable{}
)
*/

// TODO(gunnsth): Clean up, not planning to use such generic implementation as the data is not that regularly
// arranged.

/*
func unmarshal(br *byteReader, i interface{}) error {
	pv := reflect.ValueOf(i)
	if pv.Kind() != reflect.Ptr {
		return errTypeCheck
	}

	v := pv.Elem()
	if v.Kind() != reflect.Struct {
		return errTypeCheck
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tfield := v.Type().Field(i)

		sizeFieldName := tfield.Tag.Get("size")
		fmt.Printf("%s tag: %s, type: %T, size: '%s'\n", tfield.Name, tfield.Tag, field.Type, sizeFieldName)

		if !field.CanInterface() {
			continue
		}

		fint := field.Addr().Interface()
		fmt.Printf("Interface: %T\n", fint)

		switch t := fint.(type) {
		case unmarshaller:
			err := t.Unmarshal(br)
			if err != nil {
				return err
			}
			fmt.Printf("Implements unmarshaller interface!\n")
		case *uint32:
			val, err := br.readUint32()
			if err != nil {
				return err
			}
			field.SetUint(uint64(val))
		case *uint16:
			val, err := br.readUint16()
			if err != nil {
				return err
			}
			field.SetUint(uint64(val))
		case *[]uint32:
			sizeField := v.FieldByName(sizeFieldName)
			if sizeField.Kind() == 0 {
				return errors.New("size field not found")
			}
			length, err := getNumber(sizeField)
			if err != nil {
				return err
			}
			if length < 0 {
				return errRangeCheck
			}
			for i := 0; i < int(length); i++ {
				val, err := br.readUint32()
				if err != nil {
					return err
				}
				*t = append(*t, val)
			}

			fmt.Printf("Length: %d\n", length)

		default:
			fmt.Printf("Unhandled type: %T\n", t)
			return errTypeCheck
		}
	}

	return nil
}

func marshal(bw *byteWriter, v interface{}) error {
	return nil
}

func getNumber(v reflect.Value) (int64, error) {
	if !v.CanInterface() {
		return 0, errors.New("cannot interface")
	}

	switch t := v.Interface().(type) {
	case uint16:
		return int64(t), nil
	}

	return 0, errTypeCheck
}

*/
