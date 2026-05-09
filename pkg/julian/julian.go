package julian

// YMD converts a Julian day number to a year, month and day
func YMD(d uint32) (int, int, int) {
	l := d + 68569
	n := 4 * l / 146097
	l -= (146097*n + 3) / 4
	year := 4000 * (l + 1) / 1461001
	l = l - 1461*year/4 + 31
	month := 80 * l / 2447
	day := l - 2447*month/80
	l = month / 11
	month = month + 2 - 12*l
	year = 100*(n-49) + year + l
	return int(year), int(month), int(day)
}
