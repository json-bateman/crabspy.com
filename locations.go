package crabspy

type Location struct {
	Title string
	Roles []string
}

var Locations = []Location{
	{"Airplane", []string{"First Class Passenger", "Air Marshall", "Mechanic", "Air Hostess", "Co-Pilot", "Captain", "Economy Class Passenger"}},
	{"Bank", []string{"Armored Car Driver", "Manager", "Consultant", "Robber", "Security Guard", "Teller", "Customer"}},
	{"Beach", []string{"Beach Waitress", "Kite Surfer", "Lifeguard", "Thief", "Beach Photographer", "Ice Cream Truck Driver", "Beach Goer"}},
	{"Cathedral", []string{"Priest", "Beggar", "Bishop", "Tourist", "Architect", "Choir Boy", "Church Goer"}},
	{"Circus-Tent", []string{"Acrobat", "Animal Trainer", "Magician", "Ring Master", "Clown", "Juggler", "Visitor"}},
	{"Corporate-Party", []string{"DJ", "Manager", "Unwanted Guest", "CEO", "Secretary", "Delivery Boy", "Accountant"}},
	{"Crusader-Army", []string{"Monk", "Nobleman", "Servant", "Priest", "Squire", "Archer", "Knight"}},
	{"Casino", []string{"Bartender", "Head Security Guard", "Bouncer", "Manager", "Hustler", "Dealer", "Gambler"}},
	{"Day-Spa", []string{"Stylist", "Masseuse", "Manicurist", "Manager", "Receptionist", "Spa Attendant", "Customer"}},
	{"Embassy", []string{"Security Guard", "Secretary", "Ambassador", "Tourist", "Refugee", "Diplomat", "Government Official"}},
	{"Hospital", []string{"Nurse", "Doctor", "Anesthesiologist", "Intern", "Therapist", "Surgeon", "Patient"}},
	{"Hotel", []string{"Doorman", "Security Guard", "Manager", "Housekeeper", "Bartender", "Receptionist", "Customer"}},
	{"Military-Base", []string{"Deserter", "Colonel", "Medic", "Sniper", "Officer", "Tank Engineer", "Soldier"}},
	{"Movie-Studio", []string{"Stunt Man", "Sound Engineer", "Camera Man", "Director", "Costume Artist", "Producer", "Actor"}},
	{"Nightclub", []string{"DJ", "Bouncer", "Bartender", "VIP Guest", "Promoter", "VIP Guest", "Clubber"}},
	{"Ocean-Liner", []string{"Cook", "Captain", "Bartender", "Musician", "Waiter", "Mechanic", "Rich Passenger"}},
	{"Passenger-Train", []string{"Mechanic", "Border Patrol", "Train Attendant", "Restaurant Chef", "Train Driver", "Stroker", "Passenger"}},
	{"Pirate-Ship", []string{"Cook", "Slave", "Cannoneer", "Tied Up Prisoner", "Cabin Boy", "Brave Captain", "Sailor"}},
	{"Polar-Station", []string{"Medic", "Expedition Leader", "Biologist", "Radioman", "Hydrologist", "Meteorologist", "Geologist"}},
	{"Police-Station", []string{"Detective", "Lawyer", "Journalist", "Criminalist", "Archivist", "Criminal", "Patrol Officer"}},
	{"Restaurant", []string{"Musician", "Bouncer", "Hostess", "Head Chef", "Food Critic", "Waiter", "Customer"}},
	{"School", []string{"Gym Teacher", "Principal", "Security Guard", "Janitor", "Cafeteria Lady", "Maintainence Man", "Student"}},
	{"Service-Station", []string{"Manager", "Tire Specialist", "Biker", "Car Owner", "Car Wash Operator", "Electrician", "Auto Mechanic"}},
	{"Space-Station", []string{"Engineer", "Alien", "Pilot", "Commander", "Scientist", "Astronaut", "Space Tourist"}},
	{"Submarine", []string{"Cook", "Commander", "Sonar Technician", "Electronics Technician", "Radioman", "Navigator", "Sailor"}},
	{"Supermarket", []string{"Cashier", "Butcher", "Janitor", "Security Guard", "Food Sample Demonstrator", "Shelf Stocker", "Customer"}},
	{"Theater", []string{"Coat Check Lady", "Prompter", "Cashier", "Director", "Actor", "Crew Man", "Audience Member"}},
	{"University", []string{"Graduate Student", "Professor", "Dean", "Psychologist", "Maintenance Man", "Janitor", "Student"}},
	{"WWII-Squad", []string{"Resistance Fighter", "Radioman", "Scout", "Medic", "Cook", "Imprisoned Nazi", "Soldier"}},
	{"Zoo", []string{"Zookeeper", "Veterinarian", "Gift Shop Clerk", "Tour Guide", "Photographer", "Janitor", "Visitor"}},
}
