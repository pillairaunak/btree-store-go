module github.com/pillairaunak/btree-store-go

go 1.18

// Map GitHub paths to your local development paths
replace (
    github.com/pillairaunak/btree-store-go/btree => ./btree
    github.com/pillairaunak/btree-store-go/buffermanager => ./buffermanager
)
