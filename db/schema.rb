# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 2021_01_24_174240) do

  create_table "machines", force: :cascade do |t|
    t.string "hostname"
    t.string "ip_address"
    t.integer "vmid"
    t.integer "cpu"
    t.integer "ram"
    t.integer "disk"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
    t.integer "service_id"
    t.index ["service_id"], name: "index_machines_on_service_id"
  end

  create_table "services", force: :cascade do |t|
    t.string "name"
    t.integer "scale"
    t.integer "cpu"
    t.integer "ram"
    t.integer "disk"
    t.string "hostname"
    t.string "domain"
    t.datetime "created_at", precision: 6, null: false
    t.datetime "updated_at", precision: 6, null: false
  end

end