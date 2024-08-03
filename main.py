from bots.messenger.directMessage import Bot
import logging


def main():
    logging.basicConfig(
         level=logging.INFO, format='%(asctime)s [%(levelname)s]: %(message)s', datefmt='%d %B %Y %H:%M:%S')

    try:
        bot = Bot(name="Messanger Bot", url="https://www.facebook.com/")
        bot.run()
    except Exception as e:
        logging.error(f"An error occurred: {e}")


if __name__ == "__main__":
    main()