package encoding

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Avik32223/compress/internal/heap"
)

type HuffmanTreeNode struct {
	id     string
	weight int
	left   *HuffmanTreeNode
	right  *HuffmanTreeNode
}

func (h HuffmanTreeNode) Val() int {
	return h.weight
}
func (h HuffmanTreeNode) String() string {
	return fmt.Sprintf("{id: %s, weight: %d}\n", h.id, h.weight)
}

type HuffmanCode struct{}

func getFrequencyTable(s string) (frequencyTable map[rune]int) {
	frequencyTable = make(map[rune]int)
	source := []rune(s)
	for i := 0; i < len(source); i++ {
		c := source[i]
		_, ok := frequencyTable[c]
		if !ok {
			frequencyTable[c] = 0
		}
		frequencyTable[c]++
	}
	return
}

func getHuffmanCodePoints(t *HuffmanTreeNode, codePrefix string) map[string]string {
	if t == nil {
		return nil
	}
	l := t.left
	r := t.right
	if l == nil && r == nil {
		return map[string]string{
			t.id: codePrefix,
		}
	}
	m := make(map[string]string)
	lc := codePrefix + "0"
	for k, v := range getHuffmanCodePoints(l, lc) {
		m[k] = v
	}
	rc := codePrefix + "1"
	for k, v := range getHuffmanCodePoints(r, rc) {
		m[k] = v
	}
	return m
}

func encodeText(codeMap map[string]string, source string) string {
	type Result struct {
		result string
		idx    int
		s      string
	}

	var wg sync.WaitGroup
	resultCh := make(chan Result, len(source))
	for idx, i := range source {
		wg.Add(1)
		go func(idx int, i rune) {
			resultCh <- Result{
				result: codeMap[string(i)],
				idx:    idx,
				s:      string(i),
			}
			wg.Done()
		}(idx, i)
	}

	wg.Wait()
	close(resultCh)

	var results []Result
	for r := range resultCh {
		results = append(results, r)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].idx < results[j].idx
	})

	// concateneting strings using "+" was terribly slow
	// strings.Builder was amazingly performant.
	var s strings.Builder
	for _, i := range results {
		s.WriteString(i.result)
	}

	return s.String()
}

func padEncodedText(s string) string {
	pad := 8 - (len(s) % 8)
	s += strings.Repeat("0", pad)
	return s
}

func transformEncodedTextToBytes(s string) ([]byte, error) {
	i := 0
	c := make([]byte, 0)
	for i < len(s) {
		x, err := strconv.ParseUint(s[i:i+8], 2, 8)
		if err != nil {
			return nil, err
		}
		c = append(c, byte(x))
		i += 8
	}
	return c, nil
}

func getSortedFrequencyMap(m map[rune]int) (freqList []struct {
	r     rune
	count int
}) {
	freqList = make([]struct {
		r     rune
		count int
	}, 0)
	for k, v := range m {
		freqList = append(freqList,
			struct {
				r     rune
				count int
			}{
				r:     k,
				count: v,
			},
		)
	}
	sort.Slice(freqList, func(i, j int) bool {
		if freqList[i].count != freqList[j].count {
			return freqList[i].count < freqList[j].count
		}
		return freqList[i].r < freqList[j].r
	})

	return freqList
}

func (h *HuffmanCode) Encode(s string) ([]byte, error) {
	m := getFrequencyTable(s)
	if len(m) == 0 {
		return []byte{}, nil
	}
	freqList := getSortedFrequencyMap(m)
	mHeap := heap.NewMinHeap()
	for _, i := range freqList {
		a := HuffmanTreeNode{
			id:     string(i.r),
			weight: i.count,
		}
		mHeap.Insert(a)
	}

	for mHeap.Size() > 1 {
		a, _ := mHeap.Extract()
		b, _ := mHeap.Extract()
		aa := a.(HuffmanTreeNode)
		bb := b.(HuffmanTreeNode)
		c := HuffmanTreeNode{
			weight: a.Val() + b.Val(),
			left:   &aa,
			right:  &bb,
		}
		mHeap.Insert(c)
	}
	rootNode, _ := mHeap.Extract()
	rootNodeT := rootNode.(HuffmanTreeNode)
	codeMap := getHuffmanCodePoints(&rootNodeT, "")
	encodedText := encodeText(codeMap, s)
	paddedEncodedText := padEncodedText(encodedText)
	encodedData, err := transformEncodedTextToBytes(paddedEncodedText)
	if err != nil {
		return nil, err
	}

	codesInList := make([]map[string]string, 0)
	for _, v := range freqList {
		k := string(v.r)
		kv := codeMap[k]
		codesInList = append(codesInList, map[string]string{k: kv})
	}

	metadataBuf := new(bytes.Buffer)
	e := gob.NewEncoder(metadataBuf)
	err = e.Encode(codesInList)
	if err != nil {
		return nil, err
	}
	metadata := metadataBuf.Bytes()

	metadataMarker := new(bytes.Buffer)
	binary.Write(metadataMarker, binary.LittleEndian, uint64(len(metadata)))

	dataMarker := new(bytes.Buffer)
	binary.Write(dataMarker, binary.LittleEndian, uint64(len(encodedText)))

	metadata = append(metadataMarker.Bytes(), metadata...)
	encodedData = append(dataMarker.Bytes(), encodedData...)
	result := append(metadata, encodedData...)
	return result, nil
}

func (h *HuffmanCode) Decode(encodedText string) ([]byte, error) {
	encodedBytes := []byte(encodedText)
	if len(encodedBytes) == 0 {
		return []byte{}, nil
	} else if len(encodedBytes) < 8 {
		return nil, fmt.Errorf("malformed encoded text")
	}
	metadataMarker := binary.LittleEndian.Uint64(encodedBytes[:8])
	if len(encodedBytes) < 8 {
		return nil, fmt.Errorf("malformed encoded text")
	}
	encodedBytes = encodedBytes[8:]
	if len(encodedBytes) < int(metadataMarker) {
		return nil, fmt.Errorf("malformed encoded text")
	}
	encodedMetadata := encodedBytes[:metadataMarker]
	encodedBytes = encodedBytes[metadataMarker:]
	var metadata []map[string]string
	metadataBuf := bytes.NewBuffer(encodedMetadata)
	e := gob.NewDecoder(metadataBuf)
	err := e.Decode(&metadata)
	if err != nil {
		return nil, err
	}
	decodeMap := make(map[string]string)
	for _, i := range metadata {
		for k, v := range i {
			decodeMap[v] = k
		}
	}
	dataMarker := binary.LittleEndian.Uint64(encodedBytes[:8])
	if len(encodedBytes) < 8 {
		return nil, fmt.Errorf("malformed encoded text")
	}
	encodedBytes = encodedBytes[8:]

	dataMarkerCount := 0
	var result strings.Builder
	var partialString string
outerLoop:
	for _, i := range encodedBytes {
		binaryRepr := fmt.Sprintf("%08b", uint64(i))
		for i := range binaryRepr {
			dataMarkerCount++
			partialString += string(binaryRepr[i])
			if decodedValue, ok := decodeMap[partialString]; ok {
				partialString = ""
				result.WriteString(decodedValue)
			}
			if dataMarkerCount == int(dataMarker) {
				break outerLoop
			}
		}
	}
	if dataMarkerCount != int(dataMarker) {
		return nil, fmt.Errorf("malformed encoded text")
	}
	x := result.String()
	return []byte(x), nil
}
