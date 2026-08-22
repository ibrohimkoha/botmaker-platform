package templates

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

// AniTez is the anime search bot template: start, search, code lookup,
// top list and admin commands.
type AniTez struct{}

// Name returns the stable template identifier.
func (t *AniTez) Name() string { return "anitez" }

// Title returns the display name of the template.
func (t *AniTez) Title() string { return "AniTez" }

// Description returns the human readable summary used by the API.
func (t *AniTez) Description() string {
	return "Anime qidirish boti — nom yoki kod orqali anime topish"
}

// Commands lists the commands exposed by this template.
func (t *AniTez) Commands() []string {
	return []string{
		"/start", "/help",
		"/search <so'rov>",
		"/code <KOD>",
		"/top",
		"/stats",
		"/admin",
		"/broadcast <matn>",
	}
}

// Apply registers every handler of the AniTez template on the bot.
func (t *AniTez) Apply(bot *tele.Bot, opts Options) error {
	spec := Spec{
		ID:         "anitez",
		Name:       "AniTez",
		Tagline:    "Anime dunyosidagi eng yaxshi asarlarni toping: nom yoki kod bilan qidiring, inline rejimda bo'lishing!",
		SearchNoun: "anime",
		CodePrefix: "ANZ",
		Catalog:    anitezCatalog,
		ExtraCmds:  []string{"• /top — eng yuqori baholangan 10 ta anime"},
	}
	applyCore(bot, spec, opts)

	bot.Handle("/top", func(c tele.Context) error {
		_ = opts.Store.RecordSearch(opts.BotID)
		var b strings.Builder
		b.WriteString("🏆 <b>Eng yaxshi 10 ta anime:</b>\n\n")
		for i, title := range TopRated(spec.Catalog, 10) {
			fmt.Fprintf(&b, "<b>%d.</b> <b>%s</b> [<code>%s</code>] — ⭐ %.1f\n",
				i+1, esc(title.Title), title.Code, title.Rating)
		}
		b.WriteString("\nBatafsil: /code <KOD>")
		return c.Send(b.String())
	})
	return nil
}

// anitezCatalog is the searchable dataset of the AniTez bot.
var anitezCatalog = []Title{
	{
		Code: "ANZ-001", Title: "Naruto", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Shounen"},
		Year:   2002, Rating: 8.3, Episodes: 220,
		Description: "Qishlog'idagilar tomonidan chetlashtirilgan yosh ninja Naruto o'zining eng katta orzusi — Hokage bo'lish yo'lida o'sib boradi.",
		Tags:        []string{"naruto", "ninja", "hokage", "konoha"},
	},
	{
		Code: "ANZ-002", Title: "Naruto Shippuden", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Drama"},
		Year:   2007, Rating: 8.7, Episodes: 500,
		Description: "Narutoning kuchliroq bo'lib qaytishi: Sasuke bilan uchrashuv va Akatsuki tashkilotiga qarshi urush.",
		Tags:        []string{"naruto", "shippuden", "sasuke", "akatsuki"},
	},
	{
		Code: "ANZ-003", Title: "Attack on Titan", Type: TypeAnime,
		Genres: []string{"Action", "Drama", "Mystery"},
		Year:   2013, Rating: 9.0, Episodes: 87,
		Description: "Insoniyat devorlar ortida titanlardan yashirinadi. Eren Yeager o'z shahrini va onasini yo'qotgach, titanlarni yo'q qilishga qasamyod qiladi.",
		Tags:        []string{"shingeki", "titan", "eren", "levi", "aot"},
	},
	{
		Code: "ANZ-004", Title: "Death Note", Type: TypeAnime,
		Genres: []string{"Mystery", "Psychological", "Supernatural"},
		Year:   2006, Rating: 8.6, Episodes: 37,
		Description: "Ismi yozilgan har bir insonni o'ldiradigan daftar topgan daho Light Yagami o'zini Xudo deb e'lon qiladi.",
		Tags:        []string{"death note", "light", "l", "kira"},
	},
	{
		Code: "ANZ-005", Title: "One Piece", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Comedy"},
		Year:   1999, Rating: 8.9, Episodes: 1100,
		Description: "Luffy va uning ekipaji afsonaviy One Piece xazinasini topish uchun Grand Line bo'ylab sayohat qiladi.",
		Tags:        []string{"one piece", "luffy", "pirate", "kaido"},
	},
	{
		Code: "ANZ-006", Title: "Demon Slayer", Type: TypeAnime,
		Genres: []string{"Action", "Supernatural", "Historical"},
		Year:   2019, Rating: 8.7, Episodes: 55,
		Description: "Oilasi demonlar tomonidan o'ldirilgan Tanjiro singlisi Nezukoni yana odamga aylantirish uchun jang qiladi.",
		Tags:        []string{"kimetsu", "yaiba", "tanjiro", "nezuko", "slayer"},
	},
	{
		Code: "ANZ-007", Title: "Fullmetal Alchemist: Brotherhood", Type: TypeAnime,
		Genres: []string{"Action", "Fantasy", "Drama"},
		Year:   2009, Rating: 9.1, Episodes: 64,
		Description: "Aka-uka Elriclar onalarini tiriltirishga urinib hammasini yo'qotadilar va Filosofiya toshini qidirishga chiqadilar.",
		Tags:        []string{"fma", "brotherhood", "edward", "alchemy", "elric"},
	},
	{
		Code: "ANZ-008", Title: "Jujutsu Kaisen", Type: TypeAnime,
		Genres: []string{"Action", "Supernatural", "Horror"},
		Year:   2020, Rating: 8.6, Episodes: 47,
		Description: "Gojo Satoru tomonidan o'qitilayotgan yosh jujutsu sehrgarlari la'natlar bilan jang qiladi. Itadori Yuji Gojo bilan birga o'qishni boshlaydi.",
		Tags:        []string{"jujutsu", "kaisen", "gojo", "itadori", "sukuna"},
	},
	{
		Code: "ANZ-009", Title: "My Hero Academia", Type: TypeAnime,
		Genres: []string{"Action", "Superhero", "School"},
		Year:   2016, Rating: 8.4, Episodes: 138,
		Description: "Quvvatsiz tug'ilgan Izuku Midoriya dunyodagi eng buyuk qahramon bo'lish uchun All Mightning merosini qabul qiladi.",
		Tags:        []string{"boku no hero", "mha", "deku", "all might", "academia"},
	},
	{
		Code: "ANZ-010", Title: "Tokyo Ghoul", Type: TypeAnime,
		Genres: []string{"Horror", "Psychological", "Supernatural"},
		Year:   2014, Rating: 7.8, Episodes: 24,
		Description: "Kaneki Ken baxtsiz hodisadan so'ng yarim ghoulga aylanadi va ikki dunyo o'rtasida muvozanat izlaydi.",
		Tags:        []string{"tokyo ghoul", "kaneki", "ghoul", "ken"},
	},
	{
		Code: "ANZ-011", Title: "Sword Art Online", Type: TypeAnime,
		Genres: []string{"Action", "Romance", "Fantasy"},
		Year:   2012, Rating: 7.2, Episodes: 25,
		Description: "O'yin ichida qamalib qolgan minglab o'yinchilar: Kirito o'lish — o'lish degan virtual dunyoda tirik qolish uchun kurashadi.",
		Tags:        []string{"sao", "kirito", "asuna", "game", "sword"},
	},
	{
		Code: "ANZ-012", Title: "Steins;Gate", Type: TypeAnime,
		Genres: []string{"Sci-Fi", "Thriller", "Drama"},
		Year:   2011, Rating: 8.8, Episodes: 24,
		Description: "O'zini o'zi e'lon qilgan 'delusional olim' Okabe Rintaro o'tmishga xabar yubora oladigan mikroto'lqinli pech yaratadi.",
		Tags:        []string{"steins gate", "okabe", "time travel", "kurisu"},
	},
	{
		Code: "ANZ-013", Title: "Code Geass", Type: TypeAnime,
		Genres: []string{"Mecha", "Psychological", "Action"},
		Year:   2006, Rating: 8.7, Episodes: 50,
		Description: "Quvvatlangan talaba Lelouch Britanniya imperiyasiga qarshi qo'zg'olonni boshqarish uchun Geass kuchidan foydalanadi.",
		Tags:        []string{"code geass", "lelouch", "zero", "mecha"},
	},
	{
		Code: "ANZ-014", Title: "Hunter x Hunter", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Fantasy"},
		Year:   2011, Rating: 8.9, Episodes: 148,
		Description: "Gon Freecss yo'qolgan otasini topish uchun afsonaviy Hunter imtihonlaridan o'tishga harakat qiladi.",
		Tags:        []string{"hunter", "hxh", "gon", "killua"},
	},
	{
		Code: "ANZ-015", Title: "Your Name", Type: TypeAnime,
		Genres: []string{"Romance", "Drama", "Supernatural"},
		Year:   2016, Rating: 8.9,
		Description: "Shahar yigiti va qishloq qizi bir-birlarining tanalarida uyg'onishadi. Taqdir ularni kutilmagan tarzda bog'laydi.",
		Tags:        []string{"kimi no na wa", "your name", "makoto shinkai", "romance"},
	},
	{
		Code: "ANZ-016", Title: "Spirited Away", Type: TypeAnime,
		Genres: []string{"Fantasy", "Adventure", "Family"},
		Year:   2001, Rating: 8.6,
		Description: "Chihiro ruhlar olamiga tushib qoladi va ota-onasini qutqarish uchun sehrli hammomda ishlashga majbur bo'ladi.",
		Tags:        []string{"spirited away", "sen to chihiro", "miyazaki", "ghibli"},
	},
	{
		Code: "ANZ-017", Title: "One Punch Man", Type: TypeAnime,
		Genres: []string{"Action", "Comedy", "Superhero"},
		Year:   2015, Rating: 8.7, Episodes: 24,
		Description: "Saitama har qanday raqibni bitta zarba bilan yengadi — endi u o'z kuchi bilan zerikadi.",
		Tags:        []string{"one punch", "saitama", "opm", "hero"},
	},
	{
		Code: "ANZ-018", Title: "Bleach", Type: TypeAnime,
		Genres: []string{"Action", "Supernatural", "Adventure"},
		Year:   2004, Rating: 8.2, Episodes: 366,
		Description: "Ichigo Kurosaki Soul Reaper bo'lib, bo'shliqlar (Hollow) va ruhlarni himoya qilish uchun jang qiladi.",
		Tags:        []string{"bleach", "ichigo", "soul reaper", "hollow"},
	},
	{
		Code: "ANZ-019", Title: "Dragon Ball Z", Type: TypeAnime,
		Genres: []string{"Action", "Adventure", "Martial Arts"},
		Year:   1989, Rating: 8.2, Episodes: 291,
		Description: "Goku Yer sayyorasini Frieza, Cell va Buu kabi xavf-xatarlardan himoya qiladi.",
		Tags:        []string{"dragon ball", "dbz", "goku", "saiyan", "vegeta"},
	},
	{
		Code: "ANZ-020", Title: "Cowboy Bebop", Type: TypeAnime,
		Genres: []string{"Sci-Fi", "Action", "Noir"},
		Year:   1998, Rating: 8.9, Episodes: 26,
		Description: "Bounty hunterlar ekipaji — Spike va Jet — quyosh tizimi bo'ylab jinoyatchilarni ovlaydi.",
		Tags:        []string{"cowboy bebop", "spike", "jazz", "space"},
	},
	{
		Code: "ANZ-021", Title: "Vinland Saga", Type: TypeAnime,
		Genres: []string{"Action", "Historical", "Drama"},
		Year:   2019, Rating: 8.8, Episodes: 48,
		Description: "Viking yigit Thorfinn otasining qotilidan o'ch olish yo'lida jang maydonlarida o'sadi.",
		Tags:        []string{"vinland", "thorfinn", "viking", "vikinglar"},
	},
	{
		Code: "ANZ-022", Title: "Chainsaw Man", Type: TypeAnime,
		Genres: []string{"Action", "Horror", "Supernatural"},
		Year:   2022, Rating: 8.6, Episodes: 12,
		Description: "Denji iti Pochita bilan qo'shilib, demon ovchisi bo'ladi va yurak o'rnida zanjirli arra yuradi.",
		Tags:        []string{"chainsaw", "denji", "pochita", "devil"},
	},
	{
		Code: "ANZ-023", Title: "Spy x Family", Type: TypeAnime,
		Genres: []string{"Comedy", "Action", "Family"},
		Year:   2022, Rating: 8.4, Episodes: 37,
		Description: "Josus, qotil va telepat qiz — uchchala sirli a'zodan iborat soxta oila bir-birlarining sirlarini yashirishadi.",
		Tags:        []string{"spy x family", "anya", "loid", "yor"},
	},
	{
		Code: "ANZ-024", Title: "Erased", Type: TypeAnime,
		Genres: []string{"Mystery", "Thriller", "Drama"},
		Year:   2016, Rating: 8.5, Episodes: 12,
		Description: "Satoru o'zini 18 yil oldinga tashlab yuboradigan kuchga ega. Onasining qotilini topish uchun o'tmishga qaytadi.",
		Tags:        []string{"erased", "boku dake", "satoru", "time"},
	},
	{
		Code: "ANZ-025", Title: "A Silent Voice", Type: TypeAnime,
		Genres: []string{"Drama", "Romance", "School"},
		Year:   2016, Rating: 8.9,
		Description: "Bir paytlar Shoko Nishimiya karligini masxara qilgan Shoya Ishida endi o'z xatolarini yuvish yo'lini izlaydi.",
		Tags:        []string{"a silent voice", "koe no katachi", "drama", "school"},
	},
}
