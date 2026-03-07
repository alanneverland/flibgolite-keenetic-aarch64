package epub

import (
	"strings"
	"unicode"
	"path"
	"fmt"
	"github.com/vinser/flibgolite/pkg/model"
	"github.com/vinser/flibgolite/pkg/parser"
)

func (ep *OPF) GetFormat() string {
	return "epub"
}

func (ep *OPF) GetTitle() string {
	if len(ep.Metadata.Title) > 0 {
		return strings.TrimSpace(ep.Metadata.Title[0])
	}
	return ""
}

func (ep *OPF) GetSort() string {
	l := ep.Lang
	if len(ep.Metadata.Language) > 0 {
		l = ep.Metadata.Language[0]
	}
	return parser.GetSortTitle(ep.GetTitle(), parser.GetLanguageTag(l))
}

func (ep *OPF) GetYear() string {
	return parser.PickYear(ep.Metadata.Date)
}

func (ep *OPF) GetPlot() string {
	return parser.StripHTMLTags(strings.Join(ep.Metadata.Description, " "))
}

func (ep *OPF) GetCover() string {
	coverHref := ""
	
	for _, item := range ep.Manifest.Item {
		if strings.Contains(item.Properties, "cover-image") {
			coverHref = strings.TrimSpace(item.Href)
			break
		}
	}
	
	
	if coverHref == "" {
		content := ""
		for _, meta := range ep.Metadata.Meta {
			if meta.Name == "cover" {
				content = strings.TrimSpace(meta.Content)
				break
			}
		}
		if content != "" {			
			for _, item := range ep.Manifest.Item {
				if item.ID == content {
					coverHref = strings.TrimSpace(item.Href)
					break
				}
			}
		}
	}
	
	
		
	if coverHref == "" {
		for _, item := range ep.Manifest.Item {
			id := strings.ToLower(item.ID)
			href := strings.ToLower(item.Href)
			isImage := strings.HasPrefix(item.MediaType, "image/")

			if isImage && (strings.Contains(id, "cover") || strings.Contains(href, "cover")) {
				coverHref = strings.TrimSpace(item.Href)
				break
			}
		}
	}
	
	if coverHref == "" {
		for _, ref := range ep.Guide.Reference {
			if ref.Type == "cover" {
				coverHref = strings.TrimSpace(ref.Href)
				break
			}
		}
	}
	
	if coverHref != "" {		
		//fmt.Printf("DEBUG: opfPath=%s, href=%s\n", ep.opfPath, coverHref)
		return path.Join(path.Dir(ep.opfPath), coverHref)
		//return coverHref
	}
	
	return ""
}

func (ep *OPF) GetLanguage() *model.Language {
	l := ep.Lang
	if len(ep.Metadata.Language) > 0 {
		l = ep.Metadata.Language[0]
	}
	return parser.GetLanguage(l)
}

func (ep *OPF) GetAuthors() []*model.Author {
	authors := make([]*model.Author, 0)
	for _, cr := range ep.Metadata.Creator {
		a := &model.Author{}
		for _, meta := range ep.Metadata.Meta {
			if meta.Refines != "#"+cr.ID {
				continue
			}

			switch {
			case meta.Property == "role" && meta.Text == "aut":
				cr.Role = "aut"
			case meta.Property == "file-as":
				cr.FileAs = meta.Text
			}
		}
		if cr.Role == "aut" || cr.Role == "" || len(ep.Metadata.Creator) == 1 {
			parts := strings.Split(cr.Text, ",")
			name := parser.ParseFullName(parts[0])
			a.Name = strings.TrimSpace(strings.TrimSuffix(name.First+" "+name.Middle+" "+name.Last+" ("+name.Nick+")", " ()"))
			if cr.FileAs != "" {
				a.Sort = parser.AddCommaAfterLastName(parser.DelimitGluedName(cr.FileAs))
			} else {
				sortName := name.Last + ", " + name.First + " " + name.Middle + " (" + name.Nick + ")"
				a.Sort = strings.TrimSuffix(strings.TrimSpace(strings.TrimSuffix(sortName, " ()")), ",")
			}
			if len(a.Sort) > 0 {
				a.Sort = strings.ToUpper(a.Sort)
				authors = append(authors, a)
			}
		}
	}
	if len(authors) == 0 {
		authors = append(authors,
			&model.Author{
				Name: "[author not specified]",
				Sort: "[author not specified]",
			},
		)
	}
	return authors
}

func isSeparator(r rune) bool {
	return r == ',' || r == ';' || r == '-' || unicode.IsSpace(r)
}

func (ep *OPF) GetGenres() []string {
	return strings.FieldsFunc(strings.Join(ep.Metadata.Subject, " "), isSeparator)
}

func (ep *OPF) GetKeywords() string {
	return strings.Join(strings.FieldsFunc(strings.Join(ep.Metadata.Subject, " "), isSeparator), " ")
}

func (ep *OPF) GetSerie() *model.Serie {
	serie := &model.Serie{}

	// 1. Стандарт EPUB 3 (belongs-to-collection)
	for _, meta := range ep.Metadata.Meta {
		if meta.Property == "belongs-to-collection" {
			serie.Name = strings.TrimSpace(meta.Text)
			// Если у коллекции есть ID, значит у нее вероятнее всего есть и номер. 
			// Берем ее и точно останавливаем поиск.
			if meta.ID != "" {
				break
			}
		}
	}

	// 2. Расширение Calibre (calibre:series)
	if serie.Name == "" {
		for _, meta := range ep.Metadata.Meta {
			if meta.Name == "calibre:series" {
				serie.Name = strings.TrimSpace(meta.Content)
				break
			}
		}
	}

	// 3. Generic fallback (просто "series", часто бывает в конвертерах)
	if serie.Name == "" {
		for _, meta := range ep.Metadata.Meta {
			if meta.Name == "series" {
				serie.Name = strings.TrimSpace(meta.Content)
				break
			}
		}
	}

	return serie
}

func (ep *OPF) GetSerieNumber() int {
	var indexStr string
	var serieID string

	// Ищем ID основной серии для EPUB 3
	for _, meta := range ep.Metadata.Meta {
		if meta.Property == "belongs-to-collection" && meta.ID != "" {
			serieID = "#" + meta.ID
			break
		}
	}

	// 1. EPUB 3: ищем номер, который ссылается на ID серии через refines
	if serieID != "" {
		for _, meta := range ep.Metadata.Meta {
			if meta.Property == "group-position" && meta.Refines == serieID {
				indexStr = meta.Text
				break
			}
		}
	}

	// 2. Calibre: ищем тег calibre:series_index
	if indexStr == "" {
		for _, meta := range ep.Metadata.Meta {
			if meta.Name == "calibre:series_index" {
				indexStr = meta.Content
				break
			}
		}
	}

	// 3. Generic fallback: ищем просто "series_index"
	if indexStr == "" {
		for _, meta := range ep.Metadata.Meta {
			if meta.Name == "series_index" {
				indexStr = meta.Content
				break
			}
		}
	}

	// Преобразование строки в число (с поддержкой дробных серий и запятых)
	if indexStr != "" {
		// Защита от русской запятой в дробях (например, "1,5" -> "1.5")
		indexStr = strings.ReplaceAll(indexStr, ",", ".")
		
		var index float64
		// fmt.Sscanf безопасно распарсит и "2", и "2.5", и "2.0"
		fmt.Sscanf(indexStr, "%f", &index)
		return int(index)
	}

	return 0
}
