env "local" {
  src = "ent://ent/schema"
  dev = "postgres://postgres:password@localhost:5532/atlasdev?sslmode=disable"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

env "prod" {
  src = "ent://ent/schema"
  url = getenv("DATABASE_URL")

  migration {
    dir = "file://migrations"
  }
}
