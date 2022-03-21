package asset

//
//const (
//	Clickhouse Service = "clickhouse"
//	Mssql      Service = "mssql"
//	Mongodb    Service = "mongodb"
//	Mysql      Service = "mysql"
//	Postgres   Service = "postgres"
//	Cassandra  Service = "cassandra"
//	Oracle     Service = "oracle"
//	Mariadb    Service = "mariadb"
//	Redshift   Service = "redshift"
//	Couchdb    Service = "couchdb"
//	Presto     Service = "presto"
//	Optimus    Service = "optimus"
//	Grafana    Service = "grafana"
//	Metabase   Service = "metabase"
//	Tableau    Service = "tableau"
//	Superset   Service = "superset"
//	Kafka      Service = "kafka"
//)
//
//// AllSupportedServices holds a list of all supported Services struct
//var AllSupportedServices = []Service{
//	Grafana,
//	Metabase,
//	Tableau,
//	Superset,
//	Kafka,
//	Optimus,
//	Clickhouse,
//	Mssql,
//	Mongodb,
//	Mysql,
//	Postgres,
//	Cassandra,
//	Oracle,
//	Mariadb,
//	Redshift,
//	Couchdb,
//	Presto,
//}
//
//// Service specifies a supported Service name
//type Service string
//
//// String cast Service to string
//func (st Service) String() string {
//	return string(st)
//}
//
//// IsValid will validate whether the Service is valid or not
//func (st Service) IsValid() bool {
//	switch st {
//	case Grafana,
//		Metabase,
//		Tableau,
//		Superset,
//		Kafka,
//		Optimus,
//		Clickhouse,
//		Mssql,
//		Mongodb,
//		Mysql,
//		Postgres,
//		Cassandra,
//		Oracle,
//		Mariadb,
//		Redshift,
//		Couchdb,
//		Presto:
//		return true
//	}
//	return false
//}
//
//// IsServiceStringValid returns true if type string is valid/supported
////func IsServiceStringValid(ts string) bool {
////	for _, supported := range AllSupportedTypes {
////		if supported == ts {
////			return true
////		}
////	}
////	return false
////}
