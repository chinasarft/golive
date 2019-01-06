package mp4

/*
http://rk700.github.io/2018/07/26/mp4v2-vulns/
这个box应该有点特殊
如果一个box的父box是ilst类型的，那么就会认为这是一个item，并用MP4ItemAtom来保存；
如果一个类型为data的box的祖先有ilst，但父box不是ilst，那么就认为这是存储实际信息的data box，并用MP4DataAtom来保存
什么意思？
*/

type IlstBox struct {
	*Box
	SubBoxes []IBox
}
