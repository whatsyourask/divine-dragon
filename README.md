# Divine Dragon

Divine Dragon is my bachelor's degree project. It's more like a pet-project, but I have tried to move it closer towards the side where it becomes a real product. So, let's explore some features that Divine Dragon is offering to you.

## Features
---

### Modular Architecture

Divine Dragon has modular architecture (similar to Metasploit). Implement new modules easily.

>Contributors are welcome!

### Command Line Interface

As many tools have it, I decided to implement it here too.

### Do enumeration/reconnaissance remotely

Tool can perform a basic steps in enumeration process for pentesters.
* Enumeration of users via Kerberos service (borrowed some code from kerbrute).
* Enumeration of users, groups, computers, DCs, OUs, built-in groups, Kerberoasting & Asreproasting users
* Enumeration of SMB shares, including dumping all files from them if you have sufficient permissions.

### Exploit famous vulnerabilities remotely

Tool allows you to perform the following attacks:
* ASREPRoasting
* Kerberoasting
* Password Spraying via Kerberos & SMB

### Command & Control server

Inside the tool you can find a functionality to start C2 server:
* Server is implemented as a simple web server with API.
* Server will print your password to authenticate in its API (if you want that...).
* API implements Role-based authorization.

### Throw exploits locally

The previous feature was C2 server. As you can assume, Where is a C2, there's also an implant... That's basically it. You have my implant and I call it "agent". Agent is a payload that you can generate inside the CLI of the dragon. Agent is naked for now. It means that it has no evasion, no obfuscation, just simple logic of the "agent". Agent talks to the C2, asks it about new jobs. If it finds something in jobs - it will try to execute it, if there is some payload available for the job.

List of available payloads to execute through the agent:
* Enumeration via PowerShell script in the payload that uses PowerView.ps1.
* Exploit for lateral movement/privesc: Pass-the-Hash, Pass-the-Ticket.
* Post exploitation moment like DCSync attack.

All of the payloads described above are created through a basic "payload generator" module.

## Details of implementation
---

### Command Line Interface

### C2 server

### C2 & Agent communication

### Payload execution through the agent