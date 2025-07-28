use std::path::Path;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 3 {
        std::process::exit(1);
    }

    let data_dir = args[1].as_str();
    println!("data {}", data_dir);
    let git_repo_dir = args[2].as_str();
    println!("repo {}", git_repo_dir);

    let output = std::process::Command::new("git")
        .current_dir(git_repo_dir)
        .args(["rev-parse", "HEAD"])
        .output()
        .expect("failed to execute git");

    let mut commit = String::new();
    if output.status.success() {
        commit = String::from_utf8(output.stdout).unwrap();
        println!("{}", &commit);
    } else {
        eprintln!(
            "{}: {}",
            output.status,
            String::from_utf8(output.stderr).unwrap()
        );
    }

    let last = Path::new(git_repo_dir).file_name().unwrap();
    let repo_in_data = Path::new(data_dir).join(&last);

    std::fs::create_dir_all(&repo_in_data).unwrap();
    std::fs::write(repo_in_data.join(".commit"), commit).unwrap();
}
