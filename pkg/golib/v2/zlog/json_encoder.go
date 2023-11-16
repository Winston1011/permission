package zlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"permission/pkg/golib/v2/env"
)

const (
	// trace 日志前缀标识（放在[]zap.Field的第一个位置提高效率）
	TopicType = "_tp"
	// 业务日志名字
	LogNameServer = "server"
	// module 日志文件名字
	LogNameModule = "module"
)

type HookFieldFunc func(*string, []Field)

func defaultHook(_ *string, _ []Field) {

}

// RegisterZYBJSONEncoder registers a special jsonEncoder under "epoch-json" name.
func RegisterZYBJSONEncoder() error {
	return zap.RegisterEncoder("epoch-json", func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewZYBJSONEncoder(cfg), nil
	})
}

type zybJsonEncoder struct {
	zapcore.Encoder
}

func NewZYBJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	jsonEncoder := zapcore.NewJSONEncoder(cfg)
	return &zybJsonEncoder{
		Encoder: jsonEncoder,
	}
}
func (enc *zybJsonEncoder) Clone() zapcore.Encoder {
	encoderClone := enc.Encoder.Clone()
	return &zybJsonEncoder{Encoder: encoderClone}
}
func (enc *zybJsonEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 增加 trace 日志前缀，ex： tt= tp=module.log
	fName := LogNameServer
	if len(fields) > 0 && fields[0].Key == TopicType {
		fName = fields[0].String // 确保一定是string类型的
		fields = fields[1:]
	}

	logConfig.hookField(&ent.Message, fields)

	/*
		switch fName {
		case LogNameModule:
		case LogNameServer:
		default:
			// 不识别的tp修改为 server
			fName = LogNameServer
		}
	*/

	buf, err := enc.Encoder.EncodeEntry(ent, fields)
	if !env.IsDockerPlatform() || buf == nil {
		return buf, err
	}

	tt := ""
	if fName == LogNameServer {
		tt = "-notice.new"
	}
	tp := appendLogFileTail(fName, getLevelType(ent.Level))
	prefix := "tt=" + tt + " tp=" + tp + " "
	n := append([]byte(prefix), buf.Bytes()...)
	buf.Reset()
	_, _ = buf.Write(n)
	return buf, err
}

func getLevelType(lel zapcore.Level) string {
	if lel <= zapcore.InfoLevel {
		return txtLogNormal
	}
	return txtLogWarnFatal
}
