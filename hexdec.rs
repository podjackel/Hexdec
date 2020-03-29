use rand::distributions::{Distribution, Uniform};
use std::i64;
use std::io::{self, BufRead, Write};
use std::sync::Arc;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::time::{Instant};
use std::convert::TryInto;
use std::env;

fn goodbye(count: &Arc<AtomicUsize>, time: &Arc<AtomicUsize>) {
  let counts = count.load(Ordering::SeqCst);
  let mut avg = 0;
  if counts > 0 { // Handle Ctrl-C pressed before completing 1 round
    avg = time.load(Ordering::SeqCst) / counts;
  }
  let secs = avg / 1000;
  let millis = avg - (secs * 1000);
  println!("\nYou played {} iterations and your average response time was {}.{} \
    seconds. Come back again! :)", counts, secs, millis);
  std::process::exit(0);
}

fn set_max(input: &String) -> i64 {
  let output: i64;
  output = match input.parse() {
    Ok(value) => value,
    Err(_) => {
      // Handle the blank value
      if input.as_str() == "" { 256 } // Default
      else { // Not an integer
        println!("Error: {} is not a valid integer",input.as_str());
        0
      }
    },
  };
  return output;
}

fn set_mode(input: &String) -> u32 {
  let output: u32;
  output = match input.as_str() {
    "d2x" => 16,
    "x2d" => 10,
    "both" => 1,
    "" => 10,
    other => {
      println!("Error: invalid choice: {}. (choose from x2d, d2x, both)",other);
      0
    },
  };
  return output;
}

fn main() {
  let stdin = io::stdin();
  let count = Arc::new(AtomicUsize::new(0)); // Threadsafe Game counter
  let total_time = Arc::new(AtomicUsize::new(0)); // Threadsafe Elapsed Time

  print!("
     __________
    | ________ |
    ||12345678||
    |''''''''''|
    |[M|#|C][-]| HexDec - Become a Hexa(decimal) Pro!
    |[7|8|9][+]| Author: Ophir Harpaz (@ophirharpaz)
    |[4|5|6][x]| ascii art by hjw
    |[1|2|3][%]|
    |[.|O|:][=]|
    |==========|
    
    Set the game properties using:
    - the maximal number that will show up;
    - the game mode (d2x for decimal to hexa, x2d for the opposite direction, or both)

");

  let mut max: i64 = 0;
  let mut base: u32 = 0;

  let args: Vec<String> = env::args().collect();
  if args.len() > 1 {
    let mut i = 1;
    while i < args.len() {
      match args[i].as_str() {
        "--max-number" => {
          max = set_max(&args[i+1]);
          i += 1;
        },
        "--game-mode" => {
          base = set_mode(&args[i+1]);
          i += 1;
        },
        &_ => i = i,
      }
      i += 1;
    }
  }
  // Get a maximum number
  // Input validation loop
  while max == 0 {
    print!("Choose a maximal number [256]: ");
    io::stdout().flush().unwrap();
    let input = stdin.lock().lines().next().unwrap().unwrap();
    max = set_max(&input);
  }

  // Generate random numbers
  let step = Uniform::new(0x0, max);
  let mut rng = rand::thread_rng();

  // Get game mode
  // Input validation loop
  while base == 0 {
    print!("game mode [x2d]: ");
    io::stdout().flush().unwrap();
    let input  = stdin.lock().lines().next().unwrap().unwrap();
    base = set_mode(&input);
  }

  // Catch Ctrl-C and output stats
  // Clone and move the threadsafe Arc's
  let count_clone = count.clone();
  let time_clone = total_time.clone();
  ctrlc::set_handler(move|| goodbye(&count_clone, &time_clone) )
    .expect("Error setting Ctrl-C handler");

  // Start game loop
  loop {
    // Grab a new random value
    let choice = step.sample(&mut rng);
    // Random game mode "both"
    if base == 1 {
      base = match step.sample(&mut rng) % 2 {
        1 => 16,
        0 => 10,
        _ => panic!(),
      }
    }

    let start = Instant::now();
    let mut answer: i64 = 0;
    // Start a valid input loop
    'input2: loop {
      // Display the right format depending on game mode
      match base {
        16 => print!("* {} = ", choice),
        10 => print!("* 0x{:x} = ", choice),
        _ => break,
      }
      io::stdout().flush().unwrap();

      // Read user input & convert it
      let line = stdin.lock().lines().next().unwrap();
      let z = i64::from_str_radix(&line.unwrap(), base);
      // Handle conversion error e.g. "a" when x2d
      answer = match z {
          Ok(value) => value,
          Err(_) => {
              println!("invalid value, try again");
              continue;
          },
      };
      break 'input2;
    }

    // Add the elapsed time & increment game counter
    let time: usize = start.elapsed().as_millis().try_into().unwrap();
    total_time.fetch_add(time, Ordering::SeqCst);
    count.fetch_add(1, Ordering::SeqCst);

    // Check the result
    if choice == answer { continue; } 
    else { 
      println!("Oops, you got that last one wrong...");
      goodbye(&count, &total_time);
      break;
    }
  }
}
