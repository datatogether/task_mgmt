package main

// import (
// 	"database/sql"
// 	"time"
// )

// var lastUpdate time.Time

// func cron() (stop chan (bool)) {
// 	stop = make(chan (bool), 0)

// 	go func() {
// 		for {
// 			select {
// 			case <-time.Tick(time.Minute * 30):
// 				if time.Since(lastUpdate) >= time.Hour {
// 					update(appDB)
// 				}
// 			case <-stop:
// 				return
// 			}
// 		}
// 	}()

// 	return stop
// }

// func update(db *sql.DB) {
// 	lastUpdate = time.Now()
// 	t := 2
// 	done := make(chan (bool), 0)

// 	log.Info("updating:", lastUpdate)

// 	// Spin off a sources update
// 	go func() {
// 		if err := updateKiwixSources(appDB); err != nil {
// 			log.Info(err.Error())
// 		}
// 		done <- true
// 	}()

// 	// Spin off a repos update
// 	go func() {
// 		if err := updateRepos(appDB); err != nil {
// 			log.Info(err.Error())
// 		}
// 		done <- true
// 	}()

// 	for t != 0 {
// 		<-done
// 		t = t - 1
// 	}

// 	tasks, err := GenerateAvailableTasks(db)
// 	if err != nil {
// 		log.Info(err.Error())
// 	}

// 	log.Infof("generated %d tasks", len(tasks))
// }

// func updateRepos(db *sql.DB) error {
// 	repos, err := ReadRepos(db, "created DESC", 10, 0)
// 	if err != nil {
// 		return err
// 	}

// 	for _, r := range repos {
// 		commit, err := r.FetchLatestCommit()
// 		if err != nil {
// 			return err
// 		}
// 		if r.LatestCommit != commit {
// 			r.LatestCommit = commit
// 			if err := r.Save(db); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }
