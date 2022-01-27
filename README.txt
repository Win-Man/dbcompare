./dbcompare -config config/dev.toml -output=print
./dbcompare -config config/dev.toml -output=file
./dbcompare -config config/dev.toml -sql="select * from t1" -output=file