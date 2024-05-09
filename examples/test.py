import requests
import urllib3
import sys

urllib3.disable_warnings()
password = sys.argv[1]

url = 'https://10.8.0.1:8888'

# # connect agent
# connect = '/connect'
# agent_jobs = '/agent/jobs'
# new_agent = {"uuid":"79ce8549-320d-4493-bd2d-c0e9a3802e5f", "hostname": "b"*15, "username":"administrator"}
# r = requests.post(url + connect, verify=False, json=new_agent)
# print("[!] Agent connects:", r.json())
# agent_token = r.json()['token']
# print()

# login operator
login = '/login'
new_login = {"username": "c2operator", "password":password}
r = requests.post(url + login, verify=False, json=new_login)
print("[!] Operator logins:", r.json())
operator_token = r.json()['token']
print()

# operator checks active agents
operator_headers={"Authorization": "Bearer " + operator_token}
r = requests.get(url + "/operator/agents", verify=False, headers=operator_headers)
print("[!] Operator checks for agents: ", r.json())
agent_uuid = r.json()[0]['Uuid']
print() 

# operator adds a new job
# add_job = '/operator/agents/job/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_job = {"agent-uuid": agent_uuid, "job-uuid":"a"*16, "payload-filename":"reverse.exe"}
# r = requests.post(url + add_job, verify=False, headers=operator_headers, json=new_job)
# print("[!] Operator adds a new job: ", r.json())
# print()

# add_job = '/operator/agents/job/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_job = {"agent-uuid": agent_uuid, "job-uuid":"b"*16, "payload-filename":"reverse.exe"}
# r = requests.post(url + add_job, verify=False, headers=operator_headers, json=new_job)
# print("[!] Operator adds a new job: ", r.json())
# print()

# add_job = '/operator/agents/job/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_job = {"agent-uuid": agent_uuid, "job-uuid":"c"*16, "payload-filename":"reverse.exe"}
# r = requests.post(url + add_job, verify=False, headers=operator_headers, json=new_job)
# print("[!] Operator adds a new job: ", r.json())
# print()
# add_job = '/operator/agents/job/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_job = {"agent-uuid": "79ce8549-320d-4493-bd2d-c0e9a3802e5f", "job-uuid":"a"*16, "payload-filename":"bind.exe"}
# r = requests.post(url + add_job, verify=False, headers=operator_headers, json=new_job)
# print("[!] Operator adds a new job: ", r.json())
# print()
# add_job = '/operator/agents/job/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_job = {"agent-uuid": "79ce8549-320d-4493-bd2d-c0e9a3802e5f", "job-uuid":"b"*16, "payload-filename":"mimikatz.exe"}
# r = requests.post(url + add_job, verify=False, headers=operator_headers, json=new_job)
# print("[!] Operator adds a new job: ", r.json())
# print()

# operator adds a payload for a job
# add_payload = '/operator/agents/payloads/add'
# operator_headers={"Authorization": "Bearer " + operator_token}
# new_payload = {"agent-uuid": "79ce8549-320d-4493-bd2d-c0e9a3802e5f", "job-uuid":"c"*16, "payload-filename":"reverse.exe"}
# r = requests.post(url + add_payload, verify=False, headers=operator_headers, json=new_payload)
# print("[!] Operator adds a new payload for a job: ", r.json())
# print()

# operator checks active agents
operator_headers={"Authorization": "Bearer " + operator_token}
r = requests.get(url + "/operator/agents", verify=False, headers=operator_headers)
print("[!] Operator checks for agents: ", r.json())
print() 

# operator checks a job
check_job = "/operator/agents/"+agent_uuid+"/jobs"
r = requests.get(url + check_job, verify=False, headers=operator_headers)
print("[!] Operator gets jobs of specific agent: ", r.json())
print()

# agent checks a new job
# agent_headers={"Authorization": "Bearer " + agent_token}
# r = requests.get(url + agent_jobs, verify=False, headers=agent_headers)
# print("[!] Agent checks jobs: ", r.json())
# print()

# agent gets a new payload
# get_payload = '/agent/jobs/payload/' + 'c'*16
# r = requests.get(url + get_payload, verify=False, headers=agent_headers)
# print("[!] Agent gets a new payload:", r.content[:16])
# print()

# # agent gets a new payload
# get_payload = '/agent/jobs/payload/' + 'b'*16
# r = requests.get(url + get_payload, verify=False, headers=agent_headers)
# print("[!] Agent gets a new payload:", r.content[:16])
# print()

# # agent gets a new payload
# get_payload = '/agent/jobs/payload/' + 'a'*16
# r = requests.get(url + get_payload, verify=False, headers=agent_headers)
# print("[!] Agent gets a new payload:", r.content[:16])
# print()


# # agent sends a status and result of job running
# update_status = '/agent/jobs/update'
# update_job = {"job-uuid": "c" * 16, "status": True, "job-result": "Everything is fine. Here's the result."}
# r = requests.post(url +  update_status, verify=False, headers=agent_headers, json=update_job)
# print("[!] Agent updated a job:", r.json())
# print()

# operator checks a job
check_job = '/operator/agents/'+agent_uuid+'/jobs'
r = requests.get(url + check_job, verify=False, headers=operator_headers)
print("[!] Operator gets jobs of specific agent: ", r.json())
print()

# operator checks a specific job
# check_job = '/operator/agents/'+agent_uuid+'/jobs/cccccccccccccccc/status'
# r = requests.get(url + check_job, verify=False, headers=operator_headers)
# print("[!] Operator gets a specific job of specific agent: ", r.json())
# print()

# agent checks a new job
# agent_headers={"Authorization": "Bearer " + agent_token}
# r = requests.get(url + agent_jobs, verify=False, headers=agent_headers)
# print("[!] Agent checks jobs: ", r.json())
# print()

# operator checks a specific agent logs
check_logs = '/operator/agents/'+agent_uuid+'/logs'
r = requests.get(url + check_logs, verify=False, headers=operator_headers)
print("[!] Operator gets logs of specific agent: ", r.json())
print()
