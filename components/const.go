package components

const PAGE_NUM = 1
const PAGE_SIZE = 20
const TABLE_PREX = "tb_permission_"

const (
	CASBIN_RULE_PTYPE = "p"
	CASBIN_ACT_ANY    = "any"
	CASBIN_ACT_READ   = "read"
	CASBIN_ACT_WRITE  = "write"
	CASBIN_ACT_GET    = "get"
	CASBIN_ACT_POST   = "post"
)

const (
	POLICY_STATUS_ALLOW string = "allow"
	POLICY_STATUS_DENY  string = "deny"
)

const (
	GROUP_STATUS_ACTIVE  int8 = 0
	GROUP_STATUS_CLOSE   int8 = 1
	GROUP_STATUS_DELETED int8 = 9
)

const (
	USER_TYPE_INTERNAL int8 = 0
	USER_TYPE_OUTER    int8 = 1
)

const (
	NODE_TYPE_API  int8 = 0
	NODE_TYPE_PAGE int8 = 1
)
