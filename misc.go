package main
import "strings"
// NOTE(doc): sruff
type misc_struct struct{}
var misc misc_struct
func (*misc_struct)Shortener(s string, i int)string{
	if(len(s)<i){
		return s
	}else{
		return s[:i-3]+"..."
	}
}
func (*misc_struct)Replace(s string, from byte, to byte)string{
	var newS []byte
	for i:=0; i<len(s); i++{
		if(s[i]==from){
			newS=append(newS, to)
		}else{
			newS=append(newS, s[i])
		}
	}
	return string(newS)
}
func (*misc_struct)CorrectName(t string)string{
	if strings.ContainsAny(t, ":*|?<>"){
		t = misc.Replace(t, ':', '_')
		t = misc.Replace(t, '*', '_')
		t = misc.Replace(t, '|', '_')
		t = misc.Replace(t, '?', '_')
		t = misc.Replace(t, '<', '_')
		t = misc.Replace(t, '>', '_')
		return t
	}else{
		return t
	}
}