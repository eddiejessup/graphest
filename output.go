package main

import (
    "io"
    "fmt"
    "strings"
	"strconv"
    "github.com/golang/geo/r2"
)

func floatToStr(f float64) string {
    return strconv.FormatFloat(f, 'f', -1, 64)
}

func fColsToSCols(fCols []float64) []string {
    sCols := make([]string, len(fCols))
    for i := range fCols {
        sCols[i] = floatToStr(fCols[i])
    }
    return sCols
}

func bodyCols(b *Body) []string {
    var sCols []string
    for _, c := range append(b.Limbs, b.Core) {
        cFCols := []float64{c.Centre.X, c.Centre.Y, c.R}
        cSCols := fColsToSCols(cFCols)
        sCols = append(sCols, cSCols...)
    }
    return sCols
}

func writeTSVRow(wr io.Writer, cols []string) {
    row := strings.Join(cols, "\t")
    io.WriteString(wr, row + "\n")
}

func writeBody(wr io.Writer, b *Body, kind string) {
    cols := []string{kind, strconv.Itoa(len(b.Limbs))}
    cols = append(cols, bodyCols(b)...)
    writeTSVRow(wr, cols)
}

func output(wr io.Writer, movers []Mover, eggs []Egg, w r2.Point, fileName string) {
    headCols := []string{"kind", "nr_limbs"}
    circAttrs := []string{"x", "y", "R"}
    circHeadCols := make([]string, len(circAttrs))
    for j := range circAttrs {
        circHeadCols[j] = fmt.Sprintf("core_%v", circAttrs[j])
    }
    headCols = append(headCols, circHeadCols...)
    for i := range movers[0].Body.Limbs {
        for j := range circAttrs {
            circHeadCols[j] = fmt.Sprintf("limb_%v_%v", i, circAttrs[j])
        }
        headCols = append(headCols, circHeadCols...)
    }
    writeTSVRow(wr, headCols)

    for _, mover := range movers {
        writeBody(wr, &mover.Body, "m")
    }
    for _, egg := range eggs {
        writeBody(wr, &egg.Body, "e")
    }
}
