package app

/*
	业务相关配置，对应 conf/app/*.yaml (如果配置过多导致文件太大可进行拆分，不宜拆分过多)
	如果配置会因为环境不同而发生变化，且配置是业务相关的不涉及资源类配置（OP不关心），在此处设置。
	如果配置几乎不会因为环境不同而发生变化，或者基本不变化，建议直接在代码中定义即可。
	在 conf.go 中实现加载
*/

type TApp struct {
	Limit map[string]struct {
		Switch int `yaml:"switch"`
		MaxNum int `yaml:"maxNum"`
	}
}
