package main

import (
       "bufio"
       "fmt"
       "io/ioutil"
       "os"
       "regexp"
)

func main() {
     scanner := bufio.NewScanner(os.Stdin)
     for scanner.Scan() {
         includes, err := includesOfFile(scanner.Text())
         if err != nil {
             panic(err)
         }
         fmt.Printf("%s,%d\n", scanner.Text(), len(includes));
     }
     if err := scanner.Err(); err != nil {
         panic(err)
     }
}

var includeRe *regexp.Regexp = regexp.MustCompile("\\#include \"([^\"]+)\"")

func includesOfFile(filename string) ([]string, error) {
     bytes, err := ioutil.ReadFile(filename)
     if err != nil {
         return nil, err
     }

     candidates := includeRe.FindAllStringSubmatch(string(bytes), -1)

     includes := make([]string, 0, len(candidates))
     for _, match := range candidates {
         includes = append(includes, match[1])
     }
     return includes, nil
}
