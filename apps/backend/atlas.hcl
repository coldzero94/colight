env "local" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev?search_path=public"

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
