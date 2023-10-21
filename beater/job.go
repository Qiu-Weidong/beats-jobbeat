package beater

type job struct {
	Version        string         `xml:"VERSION,attr"`
	Modules        []module       `xml:"MOUDLE"`
	JobInformation jobinformation `xml:"JOBINFORMATION"`
	Plot           plot           `xml:"PLOT"`
	// VARDEFINE
}
type module struct {
	Name       string      `xml:"name,attr"`
	Version    string      `xml:"VERSION,attr"`
	Status     string      `xml:"status,attr"`
	Parameters []parameter `xml:"PARAMETER"`
}

type parameter struct {
	Name   string `xml:"name,attr"`
	Valid  string `xml:"VALID,attr"`
	Tag    string `xml:"tag,attr"`
	UiName string `xml:"uiname,attr"`
	Value  string `xml:",chardata"`
	// Value  string `xml:",chardata"`
}
type jobinformation struct {
	Project string `xml:"PROJECT"`
	Survey  string `xml:"SURVEY"`
	DbName  string `xml:"DBNAME"`
	DbUser  string `xml:"DBUSER"`
	DbPwd   string `xml:"DBPWD"`
}

type vardefine struct {
}
type plot struct {
	Poms []pom `xml:"POM"`
}
type pom struct {
	Rect rect `xml:"RECT"`
}
type rect struct {
	X int `xml:"X"`
	Y int `xml:"Y"`
}
