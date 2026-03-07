package pdf

import (
	"path/filepath"
	"strings"

	"github.com/vinser/flibgolite/pkg/model"
	"github.com/vinser/flibgolite/pkg/parser"
	"golang.org/x/text/language"
)

func (p *PDF) GetFormat() string {
	return "pdf"
}

func (p *PDF) GetTitle() string {
	title := strings.TrimSpace(p.Info["Title"])
	if title == "" {
		// Если в PDF пусто — берем имя файла
		return strings.TrimSuffix(p.FileName, filepath.Ext(p.FileName))
	}
	return title
}

func (p *PDF) GetSort() string {
	return parser.GetSortTitle(p.GetTitle(), language.English)
}

func (p *PDF) GetYear() string {
	return parser.PickYear(p.Info["CreationDate"])
}

func (p *PDF) GetPlot() string {
	return strings.TrimSpace(p.Info["Subject"])
}

func (p *PDF) GetCover() string {
	return ""
}

func (p *PDF) GetLanguage() *model.Language {
	return parser.GetLanguage("en") // Дефолтный язык для PDF
}

func (p *PDF) GetAuthors() []*model.Author {
	authors := make([]*model.Author, 0)
	rawAuthor := strings.TrimSpace(p.Info["Author"])

	if rawAuthor != "" {
		name := parser.ParseFullName(rawAuthor)
		
		// Если парсер нашел имя/фамилию — создаем объект
		if name.First != "" || name.Last != "" {
			a := &model.Author{
				Name: strings.TrimSpace(name.First + " " + name.Middle + " " + name.Last),
				Sort: strings.ToUpper(strings.TrimSuffix(name.Last+", "+name.First+" "+name.Middle, ", ")),
			}
			authors = append(authors, a)
		} else {
			// FALLBACK: Если парсер имен не справился, берем строку "как есть"
			// Это поможет для организаций или псевдонимов
			authors = append(authors, &model.Author{
				Name: rawAuthor,
				Sort: strings.ToUpper(rawAuthor),
			})
		}
	}

	if len(authors) == 0 {
		authors = append(authors, &model.Author{
			Name: "[author not specified]",
			Sort: "[author not specified]",
		})
	}
	return authors
}

func (p *PDF) GetGenres() []string {
	keywords := p.Info["Keywords"]
	if keywords == "" {
		return []string{}
	}
	return strings.FieldsFunc(keywords, func(r rune) bool {
		return r == ',' || r == ';'
	})
}

func (p *PDF) GetKeywords() string {
	return p.Info["Keywords"]
}

func (p *PDF) GetSerie() *model.Serie {
	return &model.Serie{}
}

func (p *PDF) GetSerieNumber() int {
	return 0
}