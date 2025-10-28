package infoEntropy

import (
	"log"
	"testing"
)

func TestChineseEntropy(t *testing.T) {
	e := LoadDB()
	e.UpdateUser(1, "[CQ:image,summary=&#91;动画表情&#93;,file=832188D1DE2DE13FC9AB49E0B1B0EAFE.jpg,sub_type=1,url=https://multimedia.nt.qq.com.cn/download?appid=1407&amp;fileid=EhRhUwZE4zhE_uYg-LqKRt9IwaP5dhjngw8g_woo-JGe9KfEkAMyBHByb2RQgL2jAVoQmbBegex40lgRDw7WHrBYH3oCT9yCAQJuag&amp;rkey=CAISMK2m7gd50x7p3FAnp6s1_X12MDub-97th3t6l-c22WyPOVTvAa3rJQ3Gm3DoPiwxkg,file_size=246247]")
	log.Print(e.GetEntropy(1))
}
