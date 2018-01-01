package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"jumpjump/goadb"
	"jumpjump/pictran"
	"jumpjump/utils"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/ffjson/ffjson"
)

// game start point :224 614 color is 13364,13621,15163

// JumpDirect jump direct type define
type JumpDirect string

// StaticRGBA starts
type StaticRGBA struct {
	R uint32
	G uint32
	B uint32
	A uint32
}

// JumpData jump Data For AI
type JumpData struct {
	Len       int
	PressDur  int
	Direction int
	Offset    int
	Salt      float64
}

// Common defines
var (
	CommonXRange = 720
	CommonYRange = 1280

	BlockOffset = 3 // 方块上下两个顶点的偏移量为3个像素

	TopStartPoint = &goadb.PixPoint{
		X: 0, Y: 310,
	}

	BottomSkipLen = 300 // 底部跳过的像素

	VeryCloseJumpLen = 95

	ShotPNGPath    = "pngs/screenshot.png"
	SavePNGPath    = "pngs/ret.png"
	JumpRecordPath = "JumpRecord.jump"

	Right         JumpDirect = "Right --->"
	Left          JumpDirect = "<--- Left"
	JumpDirMap    map[int]JumpDirect
	JumpDirIntMap map[JumpDirect]int

	LeftRangeBanEnd    = 320
	RightRangeBanStart = 420

	SpringGrayColorValue = 54227

	// 60909,54227,54741
	SpiritColor *StaticRGBA

	SpiritColorColorRList []*StaticRGBA
)

// set speed salt
var (
	SpeedSalt           float64
	HistoryJumpDataList []*JumpData
	JumpRecordMap       map[int][]*JumpData
)

func init() {
	SpiritColor = &StaticRGBA{
		R: 52,
		G: 53,
		B: 59,
	}

	SpiritColorColorRList = []*StaticRGBA{
		SpiritColor,
		&StaticRGBA{
			R: 52,
			G: 52,
			B: 59,
		},
		&StaticRGBA{
			R: 52,
			G: 53,
			B: 64,
		},
		&StaticRGBA{
			R: 52,
			G: 53,
			B: 61,
		},
		&StaticRGBA{
			R: 53,
			G: 54,
			B: 62,
		},
		&StaticRGBA{
			R: 52,
			G: 52,
			B: 60,
		},

		&StaticRGBA{
			R: 54,
			G: 55,
			B: 65,
		},
		&StaticRGBA{
			R: 58,
			G: 56,
			B: 71,
		},
		&StaticRGBA{
			R: 61,
			G: 56,
			B: 75,
		},

		&StaticRGBA{
			R: 64,
			G: 59,
			B: 81,
		},
		&StaticRGBA{
			R: 67,
			G: 59,
			B: 86,
		},
		&StaticRGBA{
			R: 66,
			G: 61,
			B: 89,
		},

		&StaticRGBA{
			R: 72,
			G: 66,
			B: 98,
		},
		&StaticRGBA{
			R: 73,
			G: 66,
			B: 100,
		},
		&StaticRGBA{
			R: 52,
			G: 57,
			B: 70,
		},
	}

	JumpDirMap = make(map[int]JumpDirect)
	JumpDirMap[1] = Right
	JumpDirMap[-1] = Left
	JumpDirIntMap = make(map[JumpDirect]int)
	JumpDirIntMap[Right] = 1
	JumpDirIntMap[Left] = -1

	SpeedSalt = 0.4195 // init salt
	HistoryJumpDataList = []*JumpData{}
	JumpRecordMap = make(map[int][]*JumpData)

	initLoadJumpRecordDataCache()
}

// Jump do Jump
func Jump(pressTimeDur int) {
	psFrom := &goadb.PixPoint{
		X: 100, Y: 250,
	}
	psEnd := &goadb.PixPoint{
		X: 150, Y: 250,
	}
	err := goadb.LongPress(psFrom, psEnd, pressTimeDur)
	if err != nil {
		panic(err)
	}
}

// GetDirAngle get below point
// return X:Width Y:High
func GetDirAngle(picRGBA *image.RGBA) *goadb.PixPoint {
	picB := picRGBA.Bounds()
	leftstartP := &goadb.PixPoint{
		X: LeftRangeBanEnd,
		Y: picB.Dy() - BottomSkipLen,
	}
	leftendP := &goadb.PixPoint{
		X: 0,
		Y: TopStartPoint.Y,
	}

	belowPoint := getFirstPoint(picRGBA, leftstartP, leftendP)

	rightStartP := &goadb.PixPoint{
		X: picB.Dx() - 1,
		Y: picB.Dy() - BottomSkipLen,
	}
	rightEndP := &goadb.PixPoint{
		X: RightRangeBanStart,
		Y: TopStartPoint.Y,
	}

	upperPoint := getFirstPoint(picRGBA, rightStartP, rightEndP)
	if upperPoint == nil {

	}

	log.Printf("angle below %v\tupper%v\n", belowPoint, upperPoint)

	angle := &goadb.PixPoint{
		X: upperPoint.X - belowPoint.X,
		Y: belowPoint.Y - upperPoint.Y,
	}

	return angle
}

// GetNextJumpPoint get jump point
func GetNextJumpPoint(picRGBA *image.RGBA) (*goadb.PixPoint, JumpDirect) {
	picB := picRGBA.Bounds()

	startP := &goadb.PixPoint{
		X: picB.Dx() - 1,
		Y: TopStartPoint.Y,
	}
	endP := &goadb.PixPoint{
		X: 0,
		Y: picB.Dy() - BottomSkipLen,
	}

	nextP := getFirstPoint(picRGBA, startP, endP)
	if nextP == nil {
		return nil, ""
	}

	var jdir JumpDirect
	if nextP.X > CommonXRange/2 {
		jdir = Right
	} else {
		jdir = Left
	}

	return nextP, jdir
}

// GetStartPoint get start jump point
func GetStartPoint(picRGBA *image.RGBA, jumpDir JumpDirect) *goadb.PixPoint {
	var xrStart, xrEnd int
	if jumpDir == Right {
		xrStart = 0
		xrEnd = LeftRangeBanEnd
	} else {
		xrStart = RightRangeBanStart
		xrEnd = CommonXRange
	}

	firstLinePoint := []goadb.PixPoint{}
	picB := picRGBA.Bounds()
	for yp := TopStartPoint.Y; yp < picB.Dy(); yp += 2 {
		orgRGBA := picRGBA.At(0, yp)
		for xp := xrStart; xp < xrEnd; xp += 2 {
			pColor := picRGBA.At(xp, yp)
			if isThingPoint(orgRGBA, pColor) {
				if isSpiritColor(pColor) {
					firstLinePoint = append(firstLinePoint, goadb.PixPoint{X: xp, Y: yp})
				}

			}
		}

		if len(firstLinePoint) > 0 {
			break
		}
	}

	if len(firstLinePoint) == 0 {
		return nil
	}

	endIndex := len(firstLinePoint) - 1
	startP := &goadb.PixPoint{
		X: (firstLinePoint[0].X + firstLinePoint[endIndex].X) / 2,
		Y: firstLinePoint[0].Y,
	}
	return startP
}

// GetPointOfDirectoion get point of direction
func GetPointOfDirectoion(picRGBA *image.RGBA, dirValue int) *goadb.PixPoint {
	var startP, endP *goadb.PixPoint
	picB := picRGBA.Bounds()

	if dirValue == JumpDirIntMap[Left] {
		startP = &goadb.PixPoint{
			X: 0,
			Y: TopStartPoint.Y,
		}
		endP = &goadb.PixPoint{
			X: LeftRangeBanEnd,
			Y: picB.Dy() - BottomSkipLen,
		}
	} else {
		startP = &goadb.PixPoint{
			X: picB.Dx() - 1,
			Y: TopStartPoint.Y,
		}
		endP = &goadb.PixPoint{
			X: RightRangeBanStart,
			Y: picB.Dy() - BottomSkipLen,
		}
	}

	pointRet := getFirstPoint(picRGBA, startP, endP)

	return pointRet
}

// GetJumpTimeDur get jump press time dur
func GetJumpTimeDur(startP, nextJumpP *goadb.PixPoint, salt float64) (int, int) {
	jumpLen := float64(calcAbs(startP.X - nextJumpP.X))
	timePressDur := int(jumpLen / salt)

	return int(jumpLen), timePressDur
}

// GetJumpDataAfterAdjust get jump timedur after check offset
func (jd *JumpData) GetJumpDataAfterAdjust() *JumpData {
	newSalt := jd.Salt + float64(jd.Offset)*0.0001

	td := int(float64(jd.Len) / newSalt)

	retJD := &JumpData{
		PressDur:  td,
		Offset:    0,
		Direction: jd.Direction,
		Len:       jd.Len,
		Salt:      newSalt,
	}

	return retJD
}

// RecordPreHistoryOffset record history
func RecordPreHistoryOffset(offset int) {
	num := len(HistoryJumpDataList)
	if num == 0 {
		return
	}

	HistoryJumpDataList[num-1].Offset = offset
}

// GetJumpTimeDurAI AI of jump data
// if no record and no result, return 0
func GetJumpTimeDurAI(jumpLen, direction int) *JumpData {
	var jdRet *JumpData
	jdRet = nil

	if jumpLen == 0 {
		log.Println("Get start point failed AI record ...")
	}

	if dataList, ok := JumpRecordMap[jumpLen]; ok {
		for _, perRecord := range dataList {
			if perRecord.Offset == 0 {
				log.Println("Because: offset == 0 but dir not equal", JumpDirMap[perRecord.Direction])
				log.Println("AI jump Dur: ", perRecord.PressDur)
				jdRet = perRecord
				// if i == len(dataList)-1 {
				// 	jdRet = perRecord
				// } else {

				// }
			}
		}

		if jdRet != nil {
			return jdRet
		}

		for _, perRecord := range dataList {
			if direction == perRecord.Direction {
				log.Println("Because: same dir ", JumpDirMap[perRecord.Direction])
				jdRet = perRecord.GetJumpDataAfterAdjust()
				log.Println("AI jump Dur (adjusted): ", jdRet.PressDur, "\toldTD:", perRecord.PressDur)
			}
			return jdRet
		}

		if jdRet != nil {
			return jdRet
		}

		count := len(dataList)

		log.Println("return latest one")
		jdRet = dataList[count-1].GetJumpDataAfterAdjust()
		log.Println("AI jump Dur (adjusted): ", jdRet.PressDur, "\toldTD:", dataList[count-1].PressDur)
		return jdRet
	}

	return nil
}

func main() {
	log.Println("Jump! Jump! Start")

	userSalt := float64(-1)
	autoRun := false

	for {
		jumpData := &JumpData{}
		getStartPointSucceed := false
		goadb.ScreenShot(ShotPNGPath)
		pngRetRGBA := pictran.GetPicRGBA(ShotPNGPath)
		nextJP, nextJDir := GetNextJumpPoint(pngRetRGBA)
		if nextJP == nil {
			fmt.Println("Error: get next jump point failed!")
			return
		}

		log.Printf("Next Jump P[%d:%d]\t%s\n", nextJP.X, nextJP.Y, nextJDir)
		scriptStartPoint := GetStartPoint(pngRetRGBA, nextJDir)
		if scriptStartPoint == nil {
			reckeckOK := func() bool {
				offDir := JumpDirMap[(-1)*JumpDirIntMap[nextJDir]]
				log.Println("get start point at ", nextJDir, " failed!!!!!!!!!!!!! retry directory ", offDir)
				scriptStartPoint = GetStartPoint(pngRetRGBA, offDir)
				if scriptStartPoint == nil {
					log.Println("off directory cannot found also, It perhaps too close, so set jumplen:", VeryCloseJumpLen)
					return false
				}

				log.Printf("script start point in off directory [%d:%d]\n", scriptStartPoint.X, scriptStartPoint.Y)
				nextJP = GetPointOfDirectoion(pngRetRGBA, (-1)*JumpDirIntMap[nextJDir])

				if nextJP == nil {
					fmt.Println("but In off direction is no NextJumpPoint!")
					return false
				}
				nextJDir = JumpDirMap[(-1)*JumpDirIntMap[nextJDir]]
				log.Printf("next jump point recheck result:%v\tdir:%s\n", nextJP, nextJDir)
				return true
			}()

			if reckeckOK == false {
				scriptStartPoint = &goadb.PixPoint{
					X: nextJP.X + VeryCloseJumpLen*(-1)*JumpDirIntMap[nextJDir],
					Y: nextJP.Y,
				}
			}
		} else {
			getStartPointSucceed = true
			log.Printf("script start point [%d:%d]\n", scriptStartPoint.X, scriptStartPoint.Y)
		}

		var saltValue float64
		if userSalt < 0 {
			saltValue = SpeedSalt
		} else {
			saltValue = userSalt
		}
		jumpLen, pressTD := GetJumpTimeDur(scriptStartPoint, nextJP, saltValue)
		log.Println("Jump Jump Time Press Duration:", pressTD, "len:", jumpLen, "salt:", saltValue)

		aiJData := GetJumpTimeDurAI(jumpLen, JumpDirIntMap[nextJDir])
		saltTamplate := float64(-1)
		if aiJData != nil {
			pressTD = aiJData.PressDur
			saltTamplate = SpeedSalt
			SpeedSalt = aiJData.Salt

			log.Println("**************************************")
			log.Println("AI Jump len [", aiJData.Len, "] find, using AI Dur:", aiJData.PressDur, "SpeedSalt:", aiJData.Salt)
		}

		saveRecordMap()

		if !autoRun {
			reader := bufio.NewReader(os.Stdin)
			inCodeB, _ := reader.ReadBytes('\n')
			code := strings.TrimSpace(string(inCodeB))

			if len(code) > 0 {
				if code == "save" {
					saveRecordMap()
					continue
				}

				codeField := strings.Split(code, "=")
				if len(codeField) == 2 {
					cmdK := codeField[0]
					cmdV := codeField[1]

					value, err := strconv.Atoi(cmdV)
					if err != nil {
						log.Println("Error: press time duration invalid!")
						continue
					}

					if cmdK == "t" {
						pressTD = value
					} else if cmdK == "a" { //auto run
						autoRun = true
						continue
					} else if cmdK == "o" { // offset
						offset := value
						RecordPreHistoryOffset(offset)
						saveRecordMap()
						continue
					} else if cmdK == "s" {
						if value == 0 {
							log.Println("reset salt to INIT:", SpeedSalt)
							userSalt = -1
						} else {
							if userSalt > 0 {
								userSalt = userSalt + float64(value)*0.0001
							} else {
								userSalt = SpeedSalt + float64(value)*0.0001
							}
						}

						continue
					} else {
						log.Println("error: command analyze failed!")
						continue
					}
				}

			}
		}

		jumpData.Direction = JumpDirIntMap[nextJDir]
		jumpData.Len = jumpLen
		jumpData.PressDur = pressTD
		jumpData.Salt = SpeedSalt

		HistoryJumpDataList = append(HistoryJumpDataList, jumpData)

		var recordLine []*JumpData

		if _, ok := JumpRecordMap[jumpLen]; ok {
			recordLine = JumpRecordMap[jumpLen]
		} else {
			recordLine = []*JumpData{}
		}

		if getStartPointSucceed == false {
			jumpData.Len = 0
			recordLine = []*JumpData{}
		}
		recordLine = append(recordLine, jumpData)
		JumpRecordMap[jumpData.Len] = recordLine

		Jump(pressTD)
		if saltTamplate > 0 {
			SpeedSalt = saltTamplate
		}

		time.Sleep(time.Second * 2)
	}
}

/*static funcs*/

func saveRecordMap() {
	JSONRet, err := ffjson.Marshal(JumpRecordMap)
	if err != nil {
		panic(err)
	}

	JSONContent := string(JSONRet) + "\n"

	utils.WriteFileWithLock(JumpRecordPath, []byte(JSONContent))

	log.Println("saveing succeed!")
}

func initLoadJumpRecordDataCache() {
	recordContent, err := utils.ReadFileByte(JumpRecordPath)
	if err != nil {
		panic(err)
	}

	err = ffjson.Unmarshal(recordContent, &JumpRecordMap)
	if err != nil {
		fmt.Printf("init load record map error:%v\n", err)
	}
}

func isThingPoint(origRGBA, checkRGBA color.Color) bool {
	sr, sg, sb, _ := origRGBA.RGBA()
	pr, pg, pb, _ := checkRGBA.RGBA()

	checkNum := 16

	if calcAbs(int(pr)-int(sr)) > checkNum || calcAbs(int(pg)-int(sg)) > checkNum || calcAbs(int(pb)-int(sb)) > checkNum {
		return true
	}

	return false
}

func calcAbs(a int) (ret int) {
	ret = (a ^ a>>31) - a>>31
	return
}

func isSpiritColor(checkRGBA color.Color) bool {
	pr, pg, pb, _ := checkRGBA.RGBA()

	for _, perSpiritC := range SpiritColorColorRList {
		if pr&0xFF == perSpiritC.R && pg&0xFF == perSpiritC.G && perSpiritC.B == pb&0xFF {
			return true
		}
	}

	return false
}

func colorCmp(orgRGBA, checkRGBA color.Color) bool {
	pr, pg, pb, _ := orgRGBA.RGBA()
	cr, cg, cb, _ := checkRGBA.RGBA()

	if pr == cr && pg == cg && pb == cb {
		return true
	}

	return false
}

func getFirstPoint(picRGBA *image.RGBA, starP, endP *goadb.PixPoint) *goadb.PixPoint {
	firstLinePoint := []goadb.PixPoint{}
	upper2Below := -1
	left2Right := -1

	if starP.X < endP.X {
		left2Right = 1
	}

	if starP.Y < endP.Y {
		upper2Below = 1
	}

	log.Printf("find point from %v->%v\tleft2Right=%d upper2Below=%d\n", starP, endP, left2Right, upper2Below)

	for yp := starP.Y; upper2Below*yp < upper2Below*endP.Y; yp = calcAbs(upper2Below*yp + 1) {
		firstLinePoint = []goadb.PixPoint{}
		orgRGBA := picRGBA.At(starP.X, yp)
		breaked := false
		for xp := starP.X; left2Right*xp < left2Right*endP.X; xp = calcAbs(left2Right*xp + 1) {
			pColor := picRGBA.At(xp, yp)

			// fmt.Printf("[%d:%d] ", xp, yp)
			if isThingPoint(orgRGBA, pColor) {
				if isSpiritColor(pColor) && len(firstLinePoint) == 0 {
					fmt.Println("========================================")
					break
				}

				if len(firstLinePoint) > 0 && breaked {
					firstLinePoint = []goadb.PixPoint{}
					break
				}

				firstLinePoint = append(firstLinePoint, goadb.PixPoint{X: xp, Y: yp})
			} else {
				if len(firstLinePoint) > 0 {
					breaked = true
				}
			}
		}

		if len(firstLinePoint) > 0 {
			break
		}
	}

	if len(firstLinePoint) == 0 {
		return nil
	}

	// test code

	// for _, perPoint := range firstLinePoint {
	// 	tr, tg, tb, ta := picRGBA.At(perPoint.X, perPoint.Y).RGBA()
	// 	fmt.Println("find point with color :", tr&0xFF, tg&0xFF, tb&0xFF, ta&0xFF)
	// }

	endIndex := len(firstLinePoint) - 1

	pRet := &goadb.PixPoint{
		X: (firstLinePoint[0].X + firstLinePoint[endIndex].X) / 2,
		Y: firstLinePoint[0].Y,
	}

	return pRet
}
