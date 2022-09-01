package gerrit

type SortByCreationDesc []GerritMergeRequest

func (a SortByCreationDesc) Len() int { return len(a) }
func (a SortByCreationDesc) Less(i, j int) bool {
	return a[i].CreationTimestamp.Unix() > a[j].CreationTimestamp.Unix()
}
func (a SortByCreationDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
