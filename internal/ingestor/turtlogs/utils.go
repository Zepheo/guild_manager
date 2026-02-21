package turtlogs

func mapClassId(id int) string {
	classes := map[int]string{
		1:  "Warrior",
		2:  "Paladin",
		3:  "Hunter",
		4:  "Rogue",
		5:  "Priest",
		7:  "Shaman",
		8:  "Mage",
		9:  "Warlock",
		11: "Druid",
	}

	if class, ok := classes[id]; ok {
		return class
	}
	return "Unknown"
}
