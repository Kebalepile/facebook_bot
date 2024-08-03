import os
import time
import logging
from dotenv import dotenv_values
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.chrome.service import Service as ChromeService
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from webdriver_manager.chrome import ChromeDriverManager

logging.basicConfig(level=logging.INFO)


class Bot:
    def __init__(self, name, url):
        self.name = name
        self.url = url
        self.driver = None

    def run(self):
        logging.info(f"init: {self.name}")

        # Load the .env variables
        env_vars = self.load_env()

        # Get email and password from .env
        email = env_vars["EMAIL"]
        password = env_vars["PASSWORD"]
        # Set up the Chrome driver
        chrome_options = webdriver.ChromeOptions()
        chrome_options.add_argument("--start-maximized")
        # chrome_options.add_argument("--headless")  # Remove or comment out this line to see the browser
        # Block notifications
        prefs = {"profile.default_content_setting_values.notifications": 1}
        chrome_options.add_experimental_option("prefs", prefs)
        self.driver = webdriver.Chrome(service=ChromeService(
            ChromeDriverManager().install()), options=chrome_options)

        logging.info(f"{self.name} loading URL: {self.url}")

        self.driver.get(self.url)
        self.wait_visible_and_send_keys(
            "//input[@placeholder='Email address or phone number']", email)
        self.wait_visible_and_send_keys(
            "//input[@placeholder='Password']", password)
        self.wait_visible_and_click("//button[text()='Log in']")

        # Keep the browser open and wait for user input
        self.pause(10)
        # self.wait_for_continue()

        # Handle alerts
        self.navigate_to_messenger()

    def wait_visible_and_send_keys(self, selector, keys):
        try:
            element = WebDriverWait(self.driver, 10).until(
                EC.visibility_of_element_located((By.XPATH, selector))
            )
            element.send_keys(keys)
        except Exception as e:
            logging.error(f"Error sending keys to element: {e}")

    def wait_visible_and_click(self, selector):
        try:
            element = WebDriverWait(self.driver, 10).until(
                EC.visibility_of_element_located((By.XPATH, selector))
            )
            element.click()
        except Exception as e:
            logging.error(f"Error clicking element: {e}")

    def click_svg_parent_element_by_path(self, path_value):
        try:
            nodes = self.driver.find_elements(By.CSS_SELECTOR, "svg path")
            for node in nodes:
                path_attr = node.get_attribute("d")
                if path_attr == path_value:
                    parent = node.find_element(
                        By.XPATH, "..").find_element(By.XPATH, "..")
                    if parent:
                        parent.click()
                        return
            logging.error("SVG element with the specified path not found")
        except Exception as e:
            logging.error(f"Error clicking SVG parent element: {e}")

    def click_element_with_text(self, text):
        try:
            js_code = f'''
            (function() {{
                let targetText = "{text.lower()}";
                let elements = document.body.querySelectorAll('*');
                let clicked = false;

                for (let element of elements) {{
                    if (element.textContent.toLowerCase().includes(targetText)) {{
                        let anchorElement = element.closest('a');
                        if (anchorElement) {{
                            anchorElement.click();
                            clicked = true;
                            console.log('Anchor element clicked');
                            return 'Anchor element clicked';
                        }}
                    }}
                }}

                if (!clicked) {{
                    console.log('No matching element found');
                    return 'No matching element found';
                }}
            }})();
            '''
            result = self.driver.execute_script(js_code)
            logging.info(result)
            self.find_element_and_scroll_into_view("new to the group")
        except Exception as e:
            logging.error(f"Error clicking element with text: {e}")

    def find_element_and_scroll_into_view(self, text):
        try:
            js_code = f'''
            (function() {{
                let targetText = "{text.lower()}";
                let elements = document.body.querySelectorAll('span');
                let scrolled = false;

                for (let element of elements) {{
                    if (element.textContent.toLowerCase().includes(targetText)) {{
                        element.scrollIntoView({{ behavior: 'smooth', block: 'center' }});
                        scrolled = true;
                        console.log('Element scrolled into view');
                        return 'new group members scrolled into view';
                    }}
                }}

                if (!scrolled) {{
                    console.log('No matching element found');
                    return 'No matching element found';
                }}
            }})();
            '''
            self.pause(10)
            result = self.driver.execute_script(js_code)
            logging.info(result)
        except Exception as e:
            logging.error(
                f"Error finding element and scrolling into view: {e}")

    def navigate_to_messenger(self):
        messenger_svg = "m459.603 1077.948-1.762 2.851a.89.89 0 0 1-1.302.245l-1.402-1.072a.354.354 0 0 0-.433.001l-1.893 1.465c-.253.196-.583-.112-.414-.386l1.763-2.851a.89.89 0 0 1 1.301-.245l1.402 1.072a.354.354 0 0 0 .434-.001l1.893-1.465c.253-.196.582.112.413.386M456 1073.5c-3.38 0-6 2.476-6 5.82 0 1.75.717 3.26 1.884 4.305.099.087.158.21.162.342l.032 1.067a.48.48 0 0 0 .674.425l1.191-.526a.473.473 0 0 1 .32-.024c.548.151 1.13.231 1.737.231 3.38 0 6-2.476 6-5.82 0-3.344-2.62-5.82-6-5.82"

        logging.info("Navigating to messenger")

        # First part: Click the target element
        js_code_part1 = '''
        (function() {
            const targetElements = document.querySelectorAll('.x1i10hfl.x1qjc9v5.xjbqb8w.xjqpnuy.xa49m3k.xqeqjp1.x2hbi6w.x13fuv20.xu3j5b3.x1q0q8m5.x26u7qi.x972fbf.xcfux6l.x1qhh985.xm0m39n.x9f619.x1ypdohk.xdl72j9.x2lah0s.xe8uvvx.xdj266r.x11i5rnm.xat24cr.x1mh8g0r.x2lwn1j.xeuugli.xexx8yu.x4uap5.x18d9i69.xkhd6sd.x1n2onr6.x16tdsg8.x1hl2dhg.xggy1nq.x1ja2u2z.x1t137rt.x1o1ewxj.x3x9cwd.x1e5q0jg.x13rtm0m.x1q0g3np.x87ps6o.x1lku1pv.x1a2a7pz.x1lliihq');
            if (targetElements.length) {
                targetElements[0].click();
                return 'Target element clicked';
            } else {
                return 'Target element not found';
            }
        })();
        '''

        logging.info("Clicking Messenger icon")
        # self.click_element_with_text("Messenger")
        self.click_svg_parent_element_by_path(messenger_svg)
        self.driver.execute_script(js_code_part1)
        self.pause(10)


        # Locate the contenteditable element
        contenteditable_element = self.driver.find_element(By.CSS_SELECTOR, '[contenteditable="true"]')
        # Type text into the contenteditable element
        text_message = self.end_user_message()
        escaped_text_message = text_message.replace('"', '\\"')
        logging.info("Typing end-user message into the Messenger chatbox")
        contenteditable_element.send_keys(escaped_text_message)

        # Simulate pressing the "Enter" key
        contenteditable_element.send_keys(Keys.ENTER)
        # self.wait_for_continue()
        self.wait_for_user_input()

    def end_user_message(self):
        message = input("Type your message here: ")
        return message

    def load_env(self):
        # Load environment variables from .env file
        # Reads the .env file and returns a dictionary
        env_vars = dotenv_values(".env")

        # Access variables directly from the dictionary
        email = env_vars.get("EMAIL")
        password = env_vars.get("PASSWORD")
        if not env_vars["EMAIL"] or not env_vars["PASSWORD"]:
            logging.error("Error loading .env file or missing variables")
        return env_vars

    def wait_for_continue(self):
        logging.info(
            f"{self.name} paused waiting for your command. \nType 'continue' and press Enter to proceed...")
        while True:
            user_input = input().strip()
            if user_input.lower() == "continue":
                break
            else:
                logging.info(
                    "Invalid input. Type 'continue' and press Enter to proceed...")

    def pause(self, seconds):
        logging.info(f"Pausing for {seconds} seconds...")
        time.sleep(seconds)

    def wait_for_user_input(self):
        logging.info("Type 'e' or 'exit' and press Enter to exit...")
        while True:
            user_input = input().strip()
            if user_input.lower() in ["e", "exit"]:
                break
            else:
                logging.info(
                    "Invalid input. Type 'e' or 'exit' and press Enter to exit...")


# Example usage
if __name__ == "__main__":
    bot = Bot(name="Messanger Bot", url="https://www.facebook.com/")
    bot.run()
