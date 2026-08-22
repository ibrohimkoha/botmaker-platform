package templates

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

// AniXUltra is the anime + movie search bot template: start, search,
// code lookup, category filters and admin commands.
type AniXUltra struct{}

// Name returns the stable template identifier.
func (t *AniXUltra) Name() string { return "anixultra" }

// Title returns the display name of the template.
func (t *AniXUltra) Title() string { return "AniXUltra" }

// Description returns the human readable summary used by the API.
func (t *AniXUltra) Description() string {
	return "Anime va kino qidirish boti — nom yoki kod orqali topish"
}

// Commands lists the commands exposed by this template.
func (t *AniXUltra) Commands() []string {
	return []string{
		"/start", "/help",
		"/search <so'rov>",
		"/code <KOD>",
		"/anime",
		"/kino",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the AniXUltra template on the bot.
func (t *AniXUltra) Apply(bot *tele.Bot, opts Options) error {
	spec := Spec{
		ID:         "anixultra",
		Name:       "AniXUltra",
		Tagline:    "Anime va kino dunyosining eng sara asarlari bitta botda: kod bilan toping, inline rejimda bo'lishing!",
		SearchNoun: "anime va kino",
		CodePrefix: "AXU",
		Catalog:    anixultraCatalog,
		ExtraCmds: []string{
			"• /anime — faqat animelar",
			"• /kino — faqat kino va seriallar",
		},
	}
	applyCore(bot, spec, opts)

	bot.Handle("/anime", func(c tele.Context) error {
		return sendTypeList(c, opts, spec, TypeAnime)
	})

	bot.Handle("/kino", func(c tele.Context) error {
		return sendTypeList(c, opts, spec, TypeMovie)
	})
	return nil
}

// sendTypeList replies with the best rated entries of the given type.
// The movie list also includes TV series.
func sendTypeList(c tele.Context, opts Options, spec Spec, t TitleType) error {
	_ = opts.Store.RecordSearch(opts.BotID)
	var catalog []Title
	for _, entry := range spec.Catalog {
		if entry.Type == t || (t == TypeMovie && entry.Type == TypeSeries) {
			catalog = append(catalog, entry)
		}
	}
	var b strings.Builder
	switch t {
	case TypeMovie:
		b.WriteString("🎬 <b>Eng yaxshi kino va seriallar:</b>\n\n")
	default:
		b.WriteString("🎌 <b>Eng yaxshi animelar:</b>\n\n")
	}
	for i, title := range TopRated(catalog, 10) {
		fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b> [<code>%s</code>] — ⭐ %.1f\n",
			i+1, esc(title.Title), title.Code, title.Rating)
	}
	b.WriteString("\nBatafsil: /code <KOD>")
	return c.Send(b.String())
}

// anixultraCatalog is the searchable dataset of the AniXUltra bot:
// anime, movies and TV series.
var anixultraCatalog = []Title{
	{
		Code: "AXU-001", Title: "Interstellar", Type: TypeMovie,
		Genres: []string{"Sci-Fi", "Drama", "Adventure"},
		Year:   2014, Rating: 8.7,
		Description: "Insoniyat Yerni tashlab ketishga majbur. Kuper va uning ekipaji yangi uy topish uchun qora tuynuk ortiga sayohat qiladi.",
		Tags:        []string{"interstellar", "nolan", "space", "sci-fi"},
	},
	{
		Code: "AXU-002", Title: "Inception", Type: TypeMovie,
		Genres: []string{"Sci-Fi", "Thriller", "Action"},
		Year:   2010, Rating: 8.8,
		Description: "Dominic Cobb tushlar ichiga kirib, g'oyalarni o'g'irlaydigan texnologiya bilan ishlaydi. Oxirgi topshiriq — g'oya ekish.",
		Tags:        []string{"inception", "nolan", "dream", "cobb"},
	},
	{
		Code: "AXU-003", Title: "The Dark Knight", Type: TypeMovie,
		Genres: []string{"Action", "Crime", "Drama"},
		Year:   2008, Rating: 9.0,
		Description: "Joker Gotem shahrini betartiblikka soladi va Batman o'z tamoyillarini sinovdan o'tkazadigan tanlov oldida qoladi.",
		Tags:        []string{"batman", "dark knight", "joker", "nolan"},
	},
	{
		Code: "AXU-004", Title: "The Matrix", Type: TypeMovie,
		Genres: []string{"Sci-Fi", "Action"},
		Year:   1999, Rating: 8.7,
		Description: "Haker Neo haqiqat — simulyatsiya ekanini bilib oladi va insoniyatni ozod qilish uchun Morfeus safiga qo'shiladi.",
		Tags:        []string{"matrix", "neo", "morpheus", "simulation"},
	},
	{
		Code: "AXU-005", Title: "Breaking Bad", Type: TypeSeries,
		Genres: []string{"Crime", "Drama", "Thriller"},
		Year:   2008, Rating: 9.5, Episodes: 62,
		Description: "Kimyo o'qituvchisi Walter White saraton kasalligidan so'ng oilasini ta'minlash uchun metamfetamin ishlab chiqarishni boshlaydi.",
		Tags:        []string{"breaking bad", "walter white", "heisenberg", "jesse"},
	},
	{
		Code: "AXU-006", Title: "Game of Thrones", Type: TypeSeries,
		Genres: []string{"Fantasy", "Drama", "Adventure"},
		Year:   2011, Rating: 9.2, Episodes: 73,
		Description: "Vesterosda yetti qirollik Temir taxt uchun kurashadi, shu payt shimolda qadimiy dushman uyg'onmoqda.",
		Tags:        []string{"game of thrones", "got", "westeros", "khaleesi"},
	},
	{
		Code: "AXU-007", Title: "Shawshank Redemption", Type: TypeMovie,
		Genres: []string{"Drama", "Crime"},
		Year:   1994, Rating: 9.3,
		Description: "Begunoh odam qamalgan bankir Andy Dufresne umid va do'stlik yordamida qamoqxona devorlarini yengadi.",
		Tags:        []string{"shawshank", "redemption", "prison", "andy"},
	},
	{
		Code: "AXU-008", Title: "Stranger Things", Type: TypeSeries,
		Genres: []string{"Sci-Fi", "Horror", "Mystery"},
		Year:   2016, Rating: 8.7, Episodes: 42,
		Description: "Kichik shaharchada bola g'oyib bo'ladi. Do'stlari uni qidirish chog'ida teskari dunyo (Upside Down) sirini ochadi.",
		Tags:        []string{"stranger things", "upside down", "eleven", "hawkins"},
	},
	{
		Code: "AXU-009", Title: "Naruto", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Shounen"},
		Year:   2002, Rating: 8.3, Episodes: 220,
		Description: "Yosh ninja Naruto Uzumaki Hokage bo'lish orzusini amalga oshirish uchun kurashadi.",
		Tags:        []string{"naruto", "ninja", "anime"},
	},
	{
		Code: "AXU-010", Title: "Demon Slayer", Type: TypeAnime,
		Genres: []string{"Action", "Supernatural"},
		Year:   2019, Rating: 8.7, Episodes: 55,
		Description: "Tanjiro Kamado demonlarga qarshi jangda singlisini qutqarishga intiladi.",
		Tags:        []string{"demon slayer", "kimetsu", "tanjiro", "anime"},
	},
	{
		Code: "AXU-011", Title: "Attack on Titan", Type: TypeAnime,
		Genres: []string{"Action", "Drama", "Mystery"},
		Year:   2013, Rating: 9.0, Episodes: 87,
		Description: "Eren Yeager insoniyatni titanlar xavfidan ozod qilish uchun qasamyod qiladi.",
		Tags:        []string{"attack on titan", "eren", "anime", "titan"},
	},
	{
		Code: "AXU-012", Title: "Your Name", Type: TypeAnime,
		Genres: []string{"Romance", "Drama", "Supernatural"},
		Year:   2016, Rating: 8.9,
		Description: "Tana almashinuvchi ikki yosh — Taki va Mitsuha taqdiri osmon jismlari bilan bog'lanadi.",
		Tags:        []string{"your name", "kimi no na wa", "anime", "romance"},
	},
	{
		Code: "AXU-013", Title: "The Godfather", Type: TypeMovie,
		Genres: []string{"Crime", "Drama"},
		Year:   1972, Rating: 9.2,
		Description: "Corleone mafiya oilasining patriarxi Vito o'z biznesini kichik o'g'li Maykaga topshiradi.",
		Tags:        []string{"godfather", "corleone", "mafia"},
	},
	{
		Code: "AXU-014", Title: "Pulp Fiction", Type: TypeMovie,
		Genres: []string{"Crime", "Drama"},
		Year:   1994, Rating: 8.9,
		Description: "Los-Anjeles jinoyat olamidagi bir-biriga bog'langan hikoyalar — Tarantinoning kult klassikasi.",
		Tags:        []string{"pulp fiction", "tarantino", "crime"},
	},
	{
		Code: "AXU-015", Title: "One Piece", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Comedy"},
		Year:   1999, Rating: 8.9, Episodes: 1100,
		Description: "Monkey D. Luffy va uning dengizchilari afsonaviy xazina orqasidan suzib ketishadi.",
		Tags:        []string{"one piece", "luffy", "anime", "pirate"},
	},
	{
		Code: "AXU-016", Title: "Fullmetal Alchemist: Brotherhood", Type: TypeAnime,
		Genres: []string{"Action", "Fantasy", "Drama"},
		Year:   2009, Rating: 9.1, Episodes: 64,
		Description: "Elric aka-ukalar yo'qotganlarini qaytarish uchun Filosofiya toshini qidiradilar.",
		Tags:        []string{"fma", "brotherhood", "anime", "alchemy"},
	},
	{
		Code: "AXU-017", Title: "Forrest Gump", Type: TypeMovie,
		Genres: []string{"Drama", "Romance"},
		Year:   1994, Rating: 8.8,
		Description: "Oddiy qalbli Forrest Gump Amerika tarixining eng muhim voqealari markazida bo'lib chiqadi.",
		Tags:        []string{"forrest gump", "drama"},
	},
	{
		Code: "AXU-018", Title: "Fight Club", Type: TypeMovie,
		Genres: []string{"Drama", "Thriller"},
		Year:   1999, Rating: 8.8,
		Description: "Uyqusizlikdan azob chekayotgan ofis xodimi Tyler Durden bilan tanishib, maxfiy jang klubiga a'zo bo'ladi.",
		Tags:        []string{"fight club", "tyler durden", "fincher"},
	},
	{
		Code: "AXU-019", Title: "Jujutsu Kaisen", Type: TypeAnime,
		Genres: []string{"Action", "Supernatural", "Horror"},
		Year:   2020, Rating: 8.6, Episodes: 47,
		Description: "Itadori Yuji la'natlangan barmoqni yeydi va jujutsu sehrgarlari dunyosiga qo'shiladi.",
		Tags:        []string{"jujutsu kaisen", "gojo", "anime", "sukuna"},
	},
	{
		Code: "AXU-020", Title: "Death Note", Type: TypeAnime,
		Genres: []string{"Mystery", "Psychological", "Supernatural"},
		Year:   2006, Rating: 8.6, Episodes: 37,
		Description: "Daho Light Yagami o'ldiruvchi daftar topadi va o'z adolatini o'rnatishga kirishadi.",
		Tags:        []string{"death note", "light", "anime", "kira"},
	},
	{
		Code: "AXU-021", Title: "The Lord of the Rings: The Fellowship of the Ring", Type: TypeMovie,
		Genres: []string{"Fantasy", "Adventure"},
		Year:   2001, Rating: 8.8,
		Description: "Frodo Baggins hokimiyat uzugini yo'q qilish uchun Mordorga yo'l oladi.",
		Tags:        []string{"lord of the rings", "lotr", "frodo", "fantasy"},
	},
	{
		Code: "AXU-022", Title: "Spirited Away", Type: TypeAnime,
		Genres: []string{"Fantasy", "Adventure", "Family"},
		Year:   2001, Rating: 8.6,
		Description: "Chihiro ruhlar olamida ota-onasini qutqarish yo'lini topadi.",
		Tags:        []string{"spirited away", "ghibli", "anime", "miyazaki"},
	},
	{
		Code: "AXU-023", Title: "Sherlock", Type: TypeSeries,
		Genres: []string{"Crime", "Mystery", "Drama"},
		Year:   2010, Rating: 9.1, Episodes: 13,
		Description: "Zamonaviy London detektivi Sherlock Holmes va doktor Uotson birga jinoyatlarni ochadilar.",
		Tags:        []string{"sherlock", "holmes", "cumberbatch"},
	},
	{
		Code: "AXU-024", Title: "Dark", Type: TypeSeries,
		Genres: []string{"Sci-Fi", "Mystery", "Thriller"},
		Year:   2017, Rating: 8.7, Episodes: 26,
		Description: "Kichik nemis shaharchasida yo'qolgan bolalar vaqt sayohati bilan bog'liq dahshatli sirni ochadi.",
		Tags:        []string{"dark", "time travel", "german", "netflix"},
	},
	{
		Code: "AXU-025", Title: "One Punch Man", Type: TypeAnime,
		Genres: []string{"Action", "Comedy", "Superhero"},
		Year:   2015, Rating: 8.7, Episodes: 24,
		Description: "Saitama bitta zarba bilan barchani yengadi va endi o'z kuchidan zerikadi.",
		Tags:        []string{"one punch man", "saitama", "anime", "opm"},
	},
}
