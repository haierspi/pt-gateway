package pgnotify

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// WatchTables WatchTables
func WatchTables(url string, tables []string, channelName string, callback func(id int, table string)) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal("open PostgreSQL error:" + err.Error())
	}
	defer db.Close()

	_, err = db.Exec(`
CREATE FUNCTION notify_trigger_update_or_insert_` + channelName + `() RETURNS trigger AS $$
DECLARE
BEGIN
  PERFORM pg_notify('` + channelName + `', TG_TABLE_NAME || ',' || NEW.id);
  RETURN new;
END;
$$ LANGUAGE plpgsql;
		`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Println(err.Error())
	}
	_, err = db.Exec(`
CREATE FUNCTION notify_trigger_delete_` + channelName + `() RETURNS trigger AS $$
DECLARE
BEGIN
  PERFORM pg_notify('` + channelName + `', TG_TABLE_NAME || ',' || OLD.id);
  RETURN new;
END;
$$ LANGUAGE plpgsql;
		`)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		log.Println(err.Error())
	}

	for _, table := range tables {
		query := fmt.Sprintf(
			`CREATE TRIGGER watched_table_trigger_update_or_insert_%s 
			AFTER INSERT OR UPDATE ON %s FOR EACH ROW EXECUTE PROCEDURE notify_trigger_update_or_insert_%s();`,
			channelName, table, channelName)
		_, err = db.Exec(query)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			log.Println(err.Error())
		}
	}
	for _, table := range tables {
		query := fmt.Sprintf(
			`CREATE TRIGGER watched_table_trigger_delete_%s 
			AFTER DELETE ON %s FOR EACH ROW EXECUTE PROCEDURE notify_trigger_delete_%s();`, channelName, table, channelName)
		_, err = db.Exec(query)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			log.Println(err.Error())
		}
	}

	listener := pq.NewListener(url, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Fatal("Create listener error:" + err.Error())
		}
	})
	if err := listener.Listen(channelName); err != nil {
		panic("Listen NotifyChannel error:" + err.Error())
	}
	go func() {
		for {
			select {
			case msg := <-listener.Notify:
				extras := strings.Split(msg.Extra, ",")
				tableName, idString := extras[0], extras[1]
				id, _ := strconv.Atoi(idString)
				callback(id, tableName)
			case <-time.After(time.Second * 5):
				go func() {
					listener.Ping()
				}()
			}
		}
	}()
}
