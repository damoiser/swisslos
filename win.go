package main

import (
  "encoding/csv"
  "os"
  "io"
  "strconv"
  "regexp"
  "fmt"
  "sort"
  "strings"
)

type Data struct {
  raw_entries   []Entry
  lucky_count   int
}

type Entry struct {
  date             string
  numbers          []int
  number_to_choose int
  complementary    *int
  lucky            *int
  rules            string
}

type Number_prob struct {
  number int
  times  int
  prob   float64
}
type Numbers_prob []Number_prob

func (np Numbers_prob) Len() int { return len(np) }
func (np Numbers_prob) Less(i, j int) bool { return np[i].prob < np[j].prob }
func (np Numbers_prob) Swap(i, j int) { np[i], np[j] = np[j], np[i] }

func main() {
  data := parse_file()
  data.give_my_money()
}

func parse_file() *Data {
  source_file, err := os.Open("swisslotto_numbers.csv")
  if err != nil {
    panic("cannot open source file " + err.Error())
  }
  defer source_file.Close()

  csv_reader := csv.NewReader(source_file)

  data := Data{
    raw_entries: []Entry{},
  }

  max_nr_regexp := regexp.MustCompile(`\d+\/(\d+)`)

  for {
    row, err := csv_reader.Read()
    if err != nil {
      if err == io.EOF {
        break
      }

      panic("failed to read row " + err.Error())
    }

    // skip comment lines
    if strings.HasPrefix(row[0], "#") {
      continue
    }

    numbers := []int{}
    numbers_raw := strings.Split(row[1], ",")

    for _, nr_str := range numbers_raw {
      // clean number
      nr_str = strings.TrimPrefix(nr_str, "[")
      nr_str = strings.TrimSuffix(nr_str, "]")
      nr_str = strings.Replace(nr_str, " ", "", -1)

      nr, err := strconv.Atoi(nr_str)
      if err != nil {
        panic("cannot convert number to int " + err.Error())
      }

      numbers = append(numbers, nr)
    }

    entry := Entry{
      date:    row[0],
      numbers: numbers,
      rules:   row[4],
    }

    if len(row[2]) == 0 {
      entry.complementary = nil
    } else {
      complementary, err := strconv.Atoi(row[2])
      if err != nil {
        panic("cannot convert complementary to int " + err.Error())
      }

      entry.complementary = &complementary
    }


    if len(row[3]) == 0 {
      entry.lucky = nil
    } else {
      lucky, err := strconv.Atoi(row[3])
      if err != nil {
        panic("cannot convert lucky to int " + err.Error())
      }

      entry.lucky = &lucky
    }

    max_nr_str := max_nr_regexp.FindString(row[4])
    max_nr_str_split := strings.Split(max_nr_str, "/")

    max_nr, err := strconv.Atoi(max_nr_str_split[1])
    if err != nil {
      panic("cannot parse max number " + err.Error())
    }

    entry.number_to_choose = max_nr

    data.raw_entries = append(data.raw_entries, entry)
  }

  return &data
}


// complementary omitted
func (data *Data) give_my_money() {
  stats_by_nr     := map[int]Number_prob{}
  stats_by_lucky  := map[int]Number_prob{}

  for _, entry := range data.raw_entries {

    for _, number := range entry.numbers {
      nr_prob, ok := stats_by_nr[number]
      times := 1

      if ok {
        times = nr_prob.times + 1
      }

      stats_by_nr[number] = Number_prob{
        number: number,
        times: times,
      }
    }

    if entry.lucky != nil {
      nr_prob, ok := stats_by_lucky[*entry.lucky]
      times := 1

      if ok {
        times = nr_prob.times + 1
      }

      stats_by_lucky[*entry.lucky] = Number_prob{
        number: *entry.lucky,
        times: times,
      }

      data.lucky_count++
    }
  }

  // deserialize all calculated entries and sort
  numbers   := Numbers_prob{}
  lucky_nr  := Numbers_prob{}

  for _, stat := range stats_by_nr {
    stat.prob = (float64(stat.times) / float64(len(data.raw_entries))) / 6.0

    numbers = append(numbers, stat)
  }

  for _, stat := range stats_by_lucky {
    stat.prob = float64(stat.times) / float64(data.lucky_count)

    lucky_nr = append(lucky_nr, stat)
  }

  sort.Sort(numbers)
  sort.Sort(lucky_nr)

  // print out the results
  fmt.Println("###############################")
  fmt.Println("Results are... [ entries", len(data.raw_entries),"]")
  fmt.Println("###############################")

  total_prob := 0.0
  for _, number := range numbers {
    row_mess := fmt.Sprint(number.number, "\t: ", number.prob * 100, "% ( times:", number.times, " )")

    if number.number > 40 {
      row_mess += " --> have less prob - not always present"
    }

    fmt.Println(row_mess)
    total_prob += number.prob*100
  }
  fmt.Println("-\t: ", total_prob, "%")

  fmt.Println("###############################")
  fmt.Println("Lucky numbers... [ counts", data.lucky_count, "]")
  fmt.Println("###############################")

  total_prob = 0.0
  for _, number := range lucky_nr {
    fmt.Println(number.number, "\t: ", number.prob * 100, "% ( times:", number.times, ")")
    total_prob += number.prob*100
  }
  fmt.Println("-\t: ", total_prob, "%")
}