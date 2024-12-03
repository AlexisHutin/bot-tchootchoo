package types

type Slack struct {
	APIKey  string        `yaml:"api_key"`
	Users   SlackUsers    `yaml:"users"`
	Message SlackMessages `yaml:"message"`
}

type SlackUsers struct {
	Coachs SlackCoachs `yaml:"coachs"`
}

type SlackCoachs struct {
	Men   []SlackUser `yaml:"men"`
	Women []SlackUser `yaml:"women"`
	Test  []SlackUser `yaml:"test"`
}

type SlackUser struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
}

type SlackMessages struct {
	Common SlackMessageCommon `yaml:"common"`
	Men    SlackMessageList   `yaml:"men"`
	Women  SlackMessageList   `yaml:"women"`
}

type SlackMessageCommon struct {
	Header    SlackMessageText `yaml:"header"`
	MatchInfo SlackMessageText `yaml:"info"`
	End       SlackMessageText `yaml:"end"`
	Help      SlackMessageText `yaml:"help"`
}

type SlackMessageList struct {
	List SlackMessageText `yaml:"list"`
}

type SlackMessageText struct {
	Type string `yaml:"type"`
	Text string `yaml:"text"`
}
