import json
# import psycopg2
# from psycopg2.extras import DictCursor

def message_reader(json_f):
    for line in json_f:
        yield line

# dsl = {
# 'dbname': 'db01',
# 'user': 'admin',
# 'password': '1234567',
# 'host': 'localhost',
# 'port': '54320'
# }

def get_message():
    with open('messages.json', 'r') as file:
        # cursor = pg_conn.cursor()
        file_content = file.read()
        # templates = json.dumps(json.loads(file_content))
        # mes_data =  json.dumps(json.loads(templates['mes_data']))
        # event_data = json.dumps(json.loads(templates['event_data']))
        # data = (templates['id'], templates['device_id'], templates['device_id'], templates['mes_id'], mes_data, event_data)
        # cursor.execute("call sh_ilo.set_event(%s, %s, %s, %s, %s, %s);", data)
        return file_content
        
        # gen = message_reader(file_json)
        # print(next(gen))

# get_message()

