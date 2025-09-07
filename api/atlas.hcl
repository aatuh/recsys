env "dev" {
  url = getenv("DATABASE_URL") # compose passes this env through
  migration {
    dir = "file://migrations"
    format = atlas
    revisions_schema = "public"
  }
  schemas = ["public"]
}