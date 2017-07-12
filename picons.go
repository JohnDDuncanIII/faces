package faces

import(
	"html/template"
	"os"
	"strings"
)

var path = ""
// script that parses through a picon db w/ a given email
func SearchPicons(s string) []template.HTML {
	var pBox []template.HTML
	if s == "" {
		pImg := `<img class="face" src="face/picons/misc/MISC/noface/face.gif" title="noface">`
		pBox = append(pBox, template.HTML(pImg))
	} else {
		atSign := strings.Index(s, "@")
		mfPiconDatabases := [4]string{"domains/", "users/", "misc", "usenix/"}
		count := 0
		// if we have a valid email address
		if atSign != -1 {
			host := s[atSign + 1:len(s)]
			user := s[0:atSign]
			host_pieces := strings.Split(host, ".")

			pDef := `<img class="face" src="` + path + `face/picons/unknown/` + host_pieces[len(host_pieces)-1] + `/unknown/face.gif" title="` + host_pieces[len(host_pieces)-1] + `">`
			pBox = append(pBox, template.HTML(pDef))

			for i := range mfPiconDatabases {
				piconPath := "face/picons/" + mfPiconDatabases[i] // they are stored in $PROFILEPATH$/messagefaces/picons/ by default
				if mfPiconDatabases[i] == "misc/" {
					piconPath += "MISC/"
				} // special case MISC

				// get number of database folders (probably six, but could theoretically change)
				var l = len(host_pieces)-1
				// we will check to see if we have a match at EACH depth,
				//     so keep a cloned version w/o the 'unknown/face.gif' portion
				for l >= 0 { // loop through however many pieces we have of the host
					piconPath += host_pieces[l] + "/" // add that portion of the host (ex: 'edu' or 'gettysburg' or 'cs')
					clonedLocal := piconPath
					if mfPiconDatabases[i] == "users/" {
						piconPath += user + "/"
					} else {
						piconPath += "unknown/"
					}
					piconPath += "face.gif"
					if _, err := os.Stat(piconPath); err == nil {
						if count == 0 {
							pBox[0] = template.HTML(`<img class="face" src="` + path + piconPath + `"`)
							if strings.Contains(piconPath, "users") {
								pBox[0] += template.HTML(` title="` + host_pieces[len(host_pieces)-1] + `">`)
							} else {
								pBox[0] += template.HTML(` title="` + host_pieces[l] + `">`)
							}
						} else {
							pImg := `<img class="face" src="` + path + piconPath + `"`
							if strings.Contains(piconPath, "users") {
								pImg += ` title="` + user + `">`
							} else {
								pImg += ` title="` + host_pieces[l] + `">`
							}
							pBox = append(pBox, template.HTML(pImg))
						}
						count++
					}
					piconPath = clonedLocal
					l--
				}
			}
		}
	}
	return pBox
}
